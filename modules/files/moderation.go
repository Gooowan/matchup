package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rekTypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

// Moderator checks image bytes for inappropriate content.
type Moderator interface {
	// IsSafe returns false (and a reason) if the image should be rejected.
	IsSafe(ctx context.Context, data []byte) (safe bool, reason string, err error)
}

// noopModerator always approves.
type noopModerator struct{}

func (n *noopModerator) IsSafe(_ context.Context, _ []byte) (bool, string, error) {
	return true, "", nil
}

// rekognitionModerator uses AWS Rekognition DetectModerationLabels.
type rekognitionModerator struct {
	client    *rekognition.Client
	threshold float32
	log       *slog.Logger
}

func (r *rekognitionModerator) IsSafe(ctx context.Context, data []byte) (bool, string, error) {
	out, err := r.client.DetectModerationLabels(ctx, &rekognition.DetectModerationLabelsInput{
		Image:         &rekTypes.Image{Bytes: data},
		MinConfidence: aws.Float32(r.threshold * 100),
	})
	if err != nil {
		r.log.Warn("rekognition moderation check failed", "error", err)
		return true, "", nil // fail open — don't block upload on API error
	}

	for _, label := range out.ModerationLabels {
		name := aws.ToString(label.Name)
		conf := aws.ToFloat32(label.Confidence)
		if conf >= r.threshold*100 {
			return false, fmt.Sprintf("flagged: %s (%.0f%%)", name, conf), nil
		}
	}
	return true, "", nil
}

// NewModerator constructs the moderator from env vars.
// MODERATION_PROVIDER=rekognition|disabled (default disabled)
// MODERATION_THRESHOLD=0.7 (default)
func NewModerator(log *slog.Logger) Moderator {
	provider := os.Getenv("MODERATION_PROVIDER")
	if provider != "rekognition" {
		if provider != "" && provider != "disabled" {
			log.Warn("unknown MODERATION_PROVIDER, defaulting to disabled", "value", provider)
		}
		return &noopModerator{}
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Warn("failed to load AWS config — NSFW moderation disabled", "error", err)
		return &noopModerator{}
	}

	threshold := float32(0.70)
	client := rekognition.NewFromConfig(cfg)
	log.Info("NSFW moderation enabled", "provider", "rekognition", "threshold", threshold)

	return &rekognitionModerator{client: client, threshold: threshold, log: log}
}

// ReadAndCheck reads the multipart file into memory, runs the moderation check,
// and returns a bytes.Reader suitable for re-use in upload.
func ReadAndCheck(ctx context.Context, mod Moderator, r io.Reader) (*bytes.Reader, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	data := buf.Bytes()
	safe, reason, err := mod.IsSafe(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("moderation check error: %w", err)
	}
	if !safe {
		return nil, fmt.Errorf("image rejected: %s", reason)
	}

	return bytes.NewReader(data), nil
}
