package val

import (
	"log"
	"main/util"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidations() {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		log.Fatal("failed to register custom validations")
	}
	validate.RegisterValidation("currency", validCurrency)

}
func validCurrency(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	return util.IsSupportedCurrency(currency)
}
