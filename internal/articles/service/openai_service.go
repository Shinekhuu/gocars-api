package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	articles "gocars-api/internal/articles/repository/postgresql/model"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

func GetResponseOpenAi(oem string, name string, brand string, modelOrSpecification string) (*articles.AiPart, error) {
	prompt := fmt.Sprintf(
		`Return JSON matching the provided schema.

		You are an automotive parts catalog expert.

		Part information:
		OEM: %s
		Name: %s
		Brand: %s
		Specification: %s

		Tasks:
		1. Identify the automotive part
		2. Extract specifications such as size and position
		3. Suggest vehicles that commonly use this type of part

		Rules:
		- If OEM compatibility is unknown, suggest vehicles that commonly use this specification
		- For wiper blades use size and body type
		- Do NOT invent unrealistic vehicles

		Return only JSON.`,
		oem, name, brand, modelOrSpecification,
	)

	var client = openai.NewClient()

	resp, err := client.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Model:       openai.ChatModelGPT4_1Mini,
			Temperature: openai.Float(0.1),

			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(prompt),
			},

			Text: responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
						Name: "part_schema",
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"article_product_name": map[string]interface{}{
									"type": "string",
								},
								"supplier_name": map[string]interface{}{
									"type": "string",
								},
								"oem": map[string]interface{}{
									"type": "string",
								},
								"vehicles": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"manufacturer": map[string]interface{}{
												"type": "string",
											},
											"model": map[string]interface{}{
												"type": "string",
											},
											"body_type": map[string]interface{}{
												"type": "string",
											},
											"year_from": map[string]interface{}{
												"type": "integer",
											},
											"year_to": map[string]interface{}{
												"type": "integer",
											},
										},
										"required": []string{
											"manufacturer",
											"model",
											"body_type",
											"year_from",
											"year_to",
										},
										"additionalProperties": false,
									},
								},
							},
							"required": []string{
								"article_product_name",
								"supplier_name",
								"oem",
								"vehicles",
							},
							"additionalProperties": false,
						},
					},
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	var part articles.AiPart

	err = json.Unmarshal([]byte(resp.OutputText()), &part)
	if err != nil {
		return nil, err
	}

	return &part, nil
}

func MapWithAI(oem string, name string, modelStr string, supplierName string) ([]string, error) {
	prompt := fmt.Sprintf(`
You are a TecDoc vehicle expert.

Return ONLY valid TecDoc chassis/type codes in JSON format.

INPUT:
oem: %s
name: %s
model/code: %s
brand: %s

OUTPUT
{"tecdoc_types":["ACR50"]}
`, oem, name, modelStr, supplierName)

	var client = openai.NewClient()

	resp, err := client.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Model:       openai.ChatModelGPT4_1Mini,
			Temperature: openai.Float(0),

			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(prompt),
			},

			Text: responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
						Name:   "tecdoc_mapping",
						Strict: openai.Bool(true),
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"tecdoc_types": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "string",
									},
								},
							},
							"required":             []string{"tecdoc_types"},
							"additionalProperties": false,
						},
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	raw := resp.OutputText()
	if raw == "" {
		return nil, fmt.Errorf("empty response")
	}

	var result struct {
		TecdocTypes []string `json:"tecdoc_types"`
	}

	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse error: %v | raw: %s", err, raw)
	}

	final := make([]string, 0)
	for _, t := range result.TecdocTypes {
		t = strings.ToUpper(strings.TrimSpace(t))

		if t == "" {
			continue
		}

		if strings.HasPrefix(t, "XV") {
			continue
		}

		if len(t) >= 5 {
			final = append(final, t)
		}
	}

	if len(final) > 0 {
		return final, nil
	}

	return result.TecdocTypes, nil
}
