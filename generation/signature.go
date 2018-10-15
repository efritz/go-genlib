package generation

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/efritz/go-genlib/types"
)

func GenerateFunction(methodName string, params, results []jen.Code, body ...jen.Code) jen.Code {
	return jen.Func().
		Id(methodName).
		Params(params...).
		Params(results...).
		Block(body...)
}

func GenerateMethod(receiverName, structName, methodName string, params, results []jen.Code, body ...jen.Code) jen.Code {
	return jen.Func().
		Params(jen.Id(receiverName).Op("*").Id(structName)).
		Id(methodName).
		Params(params...).
		Params(results...).
		Block(body...)
}

func GenerateOverride(receiverName, structName, importPath string, method *types.Method, body ...jen.Code) jen.Code {
	params := GenerateParamTypes(method, importPath, false)
	for i, param := range params {
		params[i] = Compose(jen.Id(fmt.Sprintf("v%d", i)), param)
	}

	return GenerateMethod(
		receiverName,
		structName,
		method.Name,
		params,
		GenerateResultTypes(method, importPath),
		body...,
	)
}

func GenerateParamTypes(method *types.Method, importPath string, omitDots bool) []jen.Code {
	params := []jen.Code{}
	for i, typ := range method.Params {
		params = append(params, GenerateType(
			typ,
			importPath,
			method.Variadic && i == len(method.Params)-1 && !omitDots,
		))
	}

	return params
}

func GenerateResultTypes(method *types.Method, importPath string) []jen.Code {
	results := []jen.Code{}
	for _, typ := range method.Results {
		results = append(results, GenerateType(
			typ,
			importPath,
			false,
		))
	}

	return results
}

func GenerateSuperCall(method *types.Method) jen.Code {
	names := []jen.Code{}
	for i := range method.Params {
		name := jen.Id(fmt.Sprintf("v%d", i))
		if method.Variadic && i == len(method.Params)-1 {
			name = Compose(name, jen.Op("..."))
		}

		names = append(names, name)
	}

	dispatch := jen.Id("m").Dot(method.Name).Call(names...)
	if len(method.Results) == 0 {
		return dispatch
	}

	assign := jen.Id("r0")
	for i := 1; i < len(method.Results); i++ {
		assign = assign.Op(",").Id(fmt.Sprintf("r%d", i))
	}

	return Compose(assign.Op(":="), dispatch)
}

func GenerateSuperReturn(method *types.Method) jen.Code {
	ret := jen.Return()

	if len(method.Results) > 0 {
		ret = ret.Id("r0")

		for i := 1; i < len(method.Results); i++ {
			ret = ret.Op(",").Id(fmt.Sprintf("r%d", i))
		}
	}

	return ret
}

// TODO - get param names
