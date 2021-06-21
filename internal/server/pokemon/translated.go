package pokemon

import (
	"context"
	"net/http"

	"github.com/jacklaaa89/pokeapi/internal/server/helpers"

	"github.com/jacklaaa89/pokeapi/internal/server/middleware"

	"github.com/jacklaaa89/pokeapi/internal/translation"
)

// Translated http.HandlerFunc which handles the /pokemon/{name}/translated endpoint.
func Translated(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	l := middleware.Logger(ctx)

	v := getVars(req)
	res, err := get(ctx, v["name"])
	if err != nil {
		l.Errorf(err.Error())
		helpers.RespondError(ctx, w, err)
		return
	}

	applyTranslation(ctx, res)
	helpers.RespondOK(ctx, w, res)
}

// applyTranslation applies the translation to the description
// if an error occurs performing the translation, the original description
// is returned.
func applyTranslation(ctx context.Context, sr *SpeciesResponse) {
	method := translation.Shakespeare
	if sr.Habitat == "cave" || sr.IsLegendary {
		method = translation.Yoda
	}

	out, err := translationAPI.Translate(ctx, sr.Description, method)
	if err == nil {
		sr.Description = out
	}
}
