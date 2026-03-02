package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func DebugPrint(format string, values ...any) {
	if gin.IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(gin.DefaultWriter, "[DEBUG] "+format, values...)
	}
}

func UUIDToString(id pgtype.UUID) string {
	var buf [36]byte

	hex.Encode(buf[0:8], id.Bytes[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], id.Bytes[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], id.Bytes[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], id.Bytes[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], id.Bytes[10:])

	return string(buf[:])
}

func StringToUUID(src string) (pgtype.UUID, error) {
	uuid := pgtype.UUID{Valid: false}

	switch len(src) {
	case 36:
		src = src[0:8] + src[9:13] + src[14:18] + src[19:23] + src[24:]
	case 32:
		// dashes already stripped, assume valid
	default:
		// assume invalid.
		return uuid, fmt.Errorf("cannot parse UUID %v", src)
	}

	decoded, err := hex.DecodeString(src)
	if err != nil || len(decoded) != 16 {
		return uuid, err
	}

	copy(uuid.Bytes[:], decoded)
	uuid.Valid = true
	return uuid, nil
}
