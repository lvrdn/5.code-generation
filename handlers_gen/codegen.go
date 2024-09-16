package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"
)

// код писать тут

type tpl struct {
	Param    string
	Cond     string
	CondText string
	Name     string
}

var (
	Tpl = template.Must(template.New("Tpl").Parse(`
	if {{.Name}} {{.Cond}} {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("{{.Param}} must be {{.CondText}}")}
	}
	`))
)

type CodegenMethodParam struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type HandlerParam struct {
	Url         string
	HandlerName string
}

type ServeHttpParam struct {
	TypeOfStruct string
	Handlers     []HandlerParam
}

func GetMapParam(str string) map[string]string {
	str = strings.ReplaceAll(str, "`", "")
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.TrimPrefix(str, "apivalidator:")
	strSlice := strings.Split(str, ",")
	mapParam := make(map[string]string, len(strSlice))
	for _, x := range strSlice {
		if strings.Contains(x, "required") {
			mapParam["required"] = ""
			continue
		}
		slice := strings.Split(x, "=")
		mapParam[slice[0]] = slice[1]
	}
	return mapParam
}

func main() {
	fileSet := token.NewFileSet()

	node, err := parser.ParseFile(fileSet, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		fmt.Println("some error happened with parse go file", err)
		return
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Println("some error happened with creafting go file", err)
		return
	}

	fmt.Fprintln(out, "package "+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "import (")
	fmt.Fprintln(out, "\t\"fmt\"")
	fmt.Fprintln(out, "\t\"encoding/json\"")
	fmt.Fprintln(out, "\t\"net/http\"")
	fmt.Fprintln(out, "\t\"strconv\"")
	fmt.Fprintln(out, "\t)")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "type RR map[string]interface{}") // мапа для организации отправки ответа

	sliceForCreateServeHTTP := []ServeHttpParam{}

	for _, f := range node.Decls {

		methodParsed, ok := f.(*ast.FuncDecl) // поиск по методам
		if ok {
			if methodParsed.Doc == nil {
				continue
			}
			comment := methodParsed.Doc.Text()
			if !strings.Contains(comment, "apigen:api") {
				continue
			}

			commentWithoutPrefix := strings.TrimPrefix(comment, "apigen:api ")

			Param := &CodegenMethodParam{}

			err := json.Unmarshal([]byte(commentWithoutPrefix), Param)

			if err != nil {
				fmt.Println(err)
				return
			}

			methodName := methodParsed.Name.Name                                                 //имя метода
			argForStructName := methodParsed.Recv.List[0].Names[0].Name                          //имя структуры метода
			typeOfStruct := (methodParsed.Recv.List[0].Type).(*ast.StarExpr).X.(*ast.Ident).Name // тип структуры
			httpArgForMethod := "w http.ResponseWriter, r *http.Request"                         //аргументы для http обертки

			flCreating := true
			for i := 0; i < len(sliceForCreateServeHTTP); i++ {
				if sliceForCreateServeHTTP[i].TypeOfStruct == typeOfStruct {
					newHandlerParam := &HandlerParam{
						Url:         Param.Url,
						HandlerName: "handler" + methodName,
					}
					sliceForCreateServeHTTP[i].Handlers = append(sliceForCreateServeHTTP[i].Handlers, *newHandlerParam)
					flCreating = false
				}
			}
			if flCreating {
				newServeHTTP := &ServeHttpParam{TypeOfStruct: typeOfStruct}
				newHandlerParam := &HandlerParam{
					Url:         Param.Url,
					HandlerName: "handler" + methodName,
				}
				newServeHTTP.Handlers = append(newServeHTTP.Handlers, *newHandlerParam)
				sliceForCreateServeHTTP = append(sliceForCreateServeHTTP, *newServeHTTP)
			}

			fmt.Fprintf(out, "func (%s *%s) handler%s(%s)"+" (", argForStructName, typeOfStruct, methodName, httpArgForMethod)

			strKey := ", "
			for i, x := range methodParsed.Type.Results.List { // типы результата функции

				if i == len(methodParsed.Type.Results.List)-1 {
					strKey = ") {\n"
				}

				if resultType, ok := (x.Type).(*ast.StarExpr); ok {
					fmt.Fprintf(out, "*%s"+strKey, resultType.X)
				} else {
					fmt.Fprintf(out, "%s"+strKey, x.Type)
				}
			}
			fmt.Fprintln(out)

			if Param.Auth {
				authCheck :=
					`
					if r.Header.Get("X-Auth") != "100500" {
					return nil, ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
					}
					`
				fmt.Fprintln(out, authCheck)
			}

			if Param.Method == "POST" {
				postCheck :=
					`
					if r.Method != "POST" {
					return nil, ApiError{http.StatusNotAcceptable, fmt.Errorf("bad method")}
					}
					`
				fmt.Fprintln(out, postCheck)
			}

			genStructName := "genStruct"                        // имя генерируемой структуры
			var ctxName string                                  //переменная контекста
			for _, arg := range methodParsed.Type.Params.List { //аргументы метода

				itemForStruct, ok := (arg.Type).(*ast.Ident)
				if ok {
					findedStruct, ok := (itemForStruct.Obj.Decl).(*ast.TypeSpec).Type.(*ast.StructType) //найденная структура
					if ok {

						fmt.Fprintf(out, "%s := &%s {}\n", genStructName, itemForStruct.Name) // создание структуры
						fmt.Fprintln(out)
						for _, field := range findedStruct.Fields.List { // по полям структуры для валидации параметров и заполнения полей

							validParams := GetMapParam(field.Tag.Value)
							x := "field" + field.Names[0].Name //переменная, которая проходит валидацию и потом записывается в поле структуры

							if value, exist := validParams["paramname"]; exist {
								fmt.Fprintf(out, "%s := r.FormValue(\"%s\")\n", x, value)
							} else {
								fmt.Fprintf(out, "%s := r.FormValue(\"%s\")\n", x, strings.ToLower(field.Names[0].Name))
							}

							if value, exist := validParams["default"]; exist {
								fmt.Fprintf(out, "if %s == \"\" {\n", x)
								fmt.Fprintf(out, " %s = \"%s\" }\n", x, value)
							}

							if field.Type.(*ast.Ident).Name == "string" {

								if _, exist := validParams["required"]; exist {
									cond := "== \"\""
									condText := "not empty"
									param := strings.ToLower(field.Names[0].Name)

									Tpl.Execute(out, tpl{param, cond, condText, x})
									fmt.Fprintln(out)
								}

								if value, exist := validParams["enum"]; exist {
									enumValueSlice := strings.Split(value, "|")
									strKey1 := ", "
									strKey2 := " && "
									cond := ""
									sliceText := ""
									for i, enumValue := range enumValueSlice {
										if i == len(enumValueSlice)-1 {
											strKey1 = ""
											strKey2 = ""
										}
										cond += x + " != " + "\"" + enumValue + "\"" + strKey2
										sliceText += enumValue + strKey1
									}

									fmt.Fprintf(out, "if %s {\n", cond)
									fmt.Fprintf(out, "return nil, ApiError{http.StatusBadRequest, fmt.Errorf(\"%s must be one of [%s]\")}\n", strings.ToLower(field.Names[0].Name), sliceText)
									fmt.Fprintf(out, "}\n")
								}

								if value, exist := validParams["min"]; exist {
									cond := " < " + value
									condText := ">= " + value
									param := strings.ToLower(field.Names[0].Name) + " len"

									Tpl.Execute(out, tpl{param, cond, condText, "len(" + x + ")"})
									fmt.Fprintln(out)
								}

								fmt.Fprintln(out)
								fmt.Fprintf(out, "%s.%s = %s\n", genStructName, field.Names[0].Name, x)
								fmt.Fprintln(out)
							}

							if field.Type.(*ast.Ident).Name == "int" {

								xInt := x + "Int"

								fmt.Fprintf(out, "%s, err := strconv.Atoi(%s)", xInt, x)
								cond := "!= nil"
								condText := "int"
								param := strings.ToLower(field.Names[0].Name)
								Tpl.Execute(out, tpl{param, cond, condText, "err"})
								fmt.Fprintln(out)

								if value, exist := validParams["min"]; exist {
									cond := "< " + value
									condText := ">= " + value
									param := strings.ToLower(field.Names[0].Name)

									Tpl.Execute(out, tpl{param, cond, condText, xInt})
									fmt.Fprintln(out)
								}

								if value, exist := validParams["max"]; exist {
									cond := "> " + value
									condText := "<= " + value
									param := strings.ToLower(field.Names[0].Name)

									Tpl.Execute(out, tpl{param, cond, condText, xInt})
									fmt.Fprintln(out)
								}
								fmt.Fprintln(out)
								fmt.Fprintf(out, "%s.%s = %s\n", genStructName, field.Names[0].Name, xInt)
							}
						}
					}
				}

				itemForCtx, ok := (arg.Type).(*ast.SelectorExpr)
				if ok {
					ctxName = arg.Names[0].Name
					fmt.Fprintf(out, "%v := r.%v()\n", ctxName, itemForCtx.Sel.Name)
					fmt.Fprintln(out)
				}
			}

			fmt.Fprintf(out, "return %v.%v(%v, *%v)\n", argForStructName, methodName, ctxName, genStructName)
			fmt.Fprintln(out, "}")
			fmt.Fprintln(out)
		}
	}

	var caseWithErrorsAndResult string = `
	
	if err != nil {

		if errorApi, ok := err.(ApiError); ok {

			w.WriteHeader(errorApi.HTTPStatus)

			dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

			w.Write(dataResutl)
			return
		}

			if err.Error() != "" {

			w.WriteHeader(http.StatusInternalServerError)

			dataResutl, _ := json.Marshal(RR{"error": err.Error()})

			w.Write(dataResutl)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	dataResult, _ := json.Marshal(
		RR{
			"error":    "",
			"response": result,
		},
	)
	w.Write(dataResult)
	`

	var defaultWithErrors string = `
	errorApi := ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}

	w.WriteHeader(errorApi.HTTPStatus)

	dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

	w.Write(dataResutl)
	`

	for i := 0; i < len(sliceForCreateServeHTTP); i++ {
		fmt.Fprintf(out, "func (h *%v) ServeHTTP(w http.ResponseWriter, r *http.Request) {\n", sliceForCreateServeHTTP[i].TypeOfStruct)
		fmt.Fprintf(out, "switch r.URL.Path {\n")

		for j := 0; j < len(sliceForCreateServeHTTP[i].Handlers); j++ {
			fmt.Fprintf(out, "case \"%v\" :\n", sliceForCreateServeHTTP[i].Handlers[j].Url)
			fmt.Fprintf(out, "result, err := h.%v(w, r)\n", sliceForCreateServeHTTP[i].Handlers[j].HandlerName)
			fmt.Fprintln(out, caseWithErrorsAndResult)
		}
		fmt.Fprintln(out, "default:")
		fmt.Fprintln(out, defaultWithErrors)

		fmt.Fprintf(out, "}\n")
		fmt.Fprintln(out)

		fmt.Fprintf(out, "}\n")
		fmt.Fprintln(out)
	}

}
