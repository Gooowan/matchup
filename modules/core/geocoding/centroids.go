package geocoding

// cityCentroids maps "country/city" (lowercased, normalized) to [lat, lng].
// Covers Ukrainian cities plus a handful of common ones as fallback.
var cityCentroids = map[string][2]float64{
	"ukraine/kyiv":          {50.4501, 30.5234},
	"ukraine/kharkiv":       {49.9935, 36.2304},
	"ukraine/odesa":         {46.4825, 30.7233},
	"ukraine/dnipro":        {48.4647, 35.0462},
	"ukraine/donetsk":       {48.0159, 37.8028},
	"ukraine/zaporizhzhia":  {47.8388, 35.1396},
	"ukraine/lviv":          {49.8397, 24.0297},
	"ukraine/kryvyi rih":    {47.9077, 33.3919},
	"ukraine/mykolaiv":      {46.9750, 31.9946},
	"ukraine/mariupol":      {47.0951, 37.5494},
	"ukraine/luhansk":       {48.5740, 39.3078},
	"ukraine/vinnytsia":     {49.2331, 28.4682},
	"ukraine/kherson":       {46.6354, 32.6169},
	"ukraine/poltava":       {49.5883, 34.5514},
	"ukraine/chernihiv":     {51.4982, 31.2893},
	"ukraine/cherkasy":      {49.4444, 32.0598},
	"ukraine/sumy":          {50.9216, 34.8003},
	"ukraine/zhytomyr":      {50.2547, 28.6587},
	"ukraine/rivne":         {50.6199, 26.2516},
	"ukraine/ivano-frankivsk": {48.9226, 24.7111},
	"ukraine/ternopil":      {49.5535, 25.5948},
	"ukraine/lutsk":         {50.7472, 25.3254},
	"ukraine/uzhhorod":      {48.6208, 22.2879},
	"ukraine/kropyvnytskyi": {48.5079, 32.2623},
	"ukraine/khmelnytskyi":  {49.4229, 26.9871},
	"ukraine/chernivtsi":    {48.2921, 25.9358},
	"ukraine/zaporizhzhya":  {47.8388, 35.1396},

	// Country-level fallbacks
	"ukraine/":  {48.3794, 31.1656},
	"poland/":   {52.2297, 21.0122},
	"germany/":  {52.5200, 13.4050},
}

// cityCentroid returns the lat/lng centroid for a given country and city.
// It tries exact match then country-only fallback.
func cityCentroid(country, city string) ([2]float64, bool) {
	key := normalizeCountry(country) + "/" + normalizeCity(city)
	if c, ok := cityCentroids[key]; ok {
		return c, true
	}
	// Country-only fallback
	countryKey := normalizeCountry(country) + "/"
	if c, ok := cityCentroids[countryKey]; ok {
		return c, true
	}
	return [2]float64{}, false
}
