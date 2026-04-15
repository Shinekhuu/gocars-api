package services

import (
	"context"
	"encoding/json"
	"fmt"
	"gocars-api/models"
	"strings"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

type TecdocResult struct {
	TecdocTypes []string `json:"tecdoc_types"`
}

func GetResponseOpenAi(
	oem string,
	name string,
	brand string,
	modelOrSpecification string,
) (*models.AiPart, error) {

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

	var part models.AiPart

	err = json.Unmarshal([]byte(resp.OutputText()), &part)
	if err != nil {
		return nil, err
	}

	return &part, nil
}

// ---- OpenAI Mapping ----
func MapWithAI(oem string, name string, modelStr string, supplierName string) ([]string, error) {
	prompt := fmt.Sprintf(`
You are a TecDoc vehicle expert.

Return ONLY valid TecDoc chassis/type codes in JSON format.

---------------------
INPUT FORMAT
---------------------
You will receive structured data:

- name: product name (LOW priority, use only if it contains model info)
- oem: OEM number (SECONDARY signal)
- brand: LOW priority (use only for disambiguation)
- model/code: PRIMARY and HIGHEST priority

INPUT:
oem: %s
name: %s
model/code: %s
brand: %s

---------------------
PRIORITY (STRICT)
---------------------
1. model/code  → ALWAYS use first
2. oem         → use ONLY if model/code is unclear
3. name/brand  → LAST fallback only

- If engine is detected:
  DO NOT rely on name or brand

---------------------
DEFINITION
---------------------
Valid TecDoc chassis:
- Format: 2–5 uppercase letters + 2–3 digits (+ optional letter)
- Wildcard allowed: e.g. TRJ12#, GRS20#

---------------------
PROCESS
---------------------
1. Normalize (uppercase, remove symbols)
2. Extract tokens from model/code
3. Classify tokens:
   - chassis → use directly
   - wildcard (#) → keep as-is
   - engine (e.g. 1AR, 2AR, 2TR, 8AR, 2AZ-FE, 1ZR, 2ZR, 3ZR) → MUST map to chassis
   - model (e.g. RX350, CAMRY, 皇冠) → filter results
   - invalid → ignore
4. Apply Chinese mapping (fixed only, no guessing)

---------------------
CORE LOGIC
---------------------
- If chassis exists → return it

- If only engines:
  → MUST convert ALL engines to chassis
  → return MOST COMMON compatible chassis
  → limit 1–2 per engine
  → NEVER return engine codes

- If engine + model:
  → BOTH must match
  → engine match is MANDATORY
  → model is FILTER ONLY

- If multiple models:
  → merge valid results only

---------------------
ENGINE RULE (STRICT)
---------------------
- NEVER return engine codes
- Engines MUST be converted to chassis

🚨 CRITICAL:
- If output contains engine-like patterns (e.g. *ZR, *AR, *GR, *TR)
  → INVALID OUTPUT
  → MUST convert to chassis before returning

- Engine tokens must NEVER appear in final output

---------------------
🚨 ENGINE COMPATIBILITY (CRITICAL)
---------------------
- DO NOT mix engine families
- If engine is present:
  → ALL returned chassis MUST match that engine family
  → If NOT → REJECT

Examples:
- 2GR → ONLY V6 chassis (GSE*, GRS*, GGL*)
- 8AR → ONLY (ARS*, ASE*, ARL*)
- 1ZR/2ZR → ZRE*
- 3ZR → ZRE*

---------------------
ENGINE → CHASSIS MAPPING (HARD RULE)
---------------------

Use the following mapping as source of truth:

1ZR → ZRE15#
2ZR → ZRE18#
3ZR → ZRE17#

1AR → GGL15
2AR → ASV50, AVV50
4AR → ASV50
8AR → ARS210, ASE30, ARL10

2GR → GSE2#, GRS19#, GGL15
1GR → GRJ15#

2TR → TRJ12#

5VZ → VZJ9#

1UR → USF40

Rules:
- MUST map engines using this table
- Mapping table OVERRIDES all other logic
- DO NOT invent new mappings
- DO NOT infer outside the table
- If engine not in table → return empty

---------------------
🚨 VEHICLE CLASS RULE (STRICT)
---------------------
- NEVER mix vehicle classes:

  sedan ≠ SUV ≠ minivan ≠ pickup

- Model (from model/code) defines PRIMARY vehicle class

- OEM MUST NOT override model-derived vehicle class

- If OEM mapping conflicts with model class:
  → REJECT the OEM-derived chassis

- NEVER return chassis from a different vehicle class

Examples:
- RAV4 (SUV) → MUST NOT return ACR*, ANH*, GGH* (minivan)
- IS / GS (sedan) → MUST NOT return SUV or minivan chassis
- ALPHARD (minivan) → MUST NOT return sedan chassis

Fallback:
- If model is generic but OEM is available:
  → USE OEM only if class is consistent

- If conflict remains:
  → RETURN empty array

---------------------
WILDCARD RULE
---------------------
- Keep wildcard EXACTLY as-is
- DO NOT expand "#"
- Prefer wildcard over long lists

---------------------
CHINESE MAPPING
---------------------
埃尔法=ALPHARD
卡罗拉=COROLLA
花冠=COROLLA
威驰=VIOS
凯美瑞=CAMRY
汉兰达=HIGHLANDER
皇冠=CROWN

---------------------
STRICT RULES
---------------------
- Return ONLY chassis codes
- No engines, no models, no explanation
- Ignore invalid tokens (e.g. TRB)
- DO NOT guess unknown mappings
- DO NOT mix unrelated families
- Prefer wildcard representation

- If conflict exists:
  → DROP invalid results
  → RETURN only valid ones
  → If none valid → return empty array
  
- NEVER concatenate tokens
- Engine/displacement numbers (e.g. 2700, 3400) MUST NOT be appended to chassis
- Chassis codes must remain atomic

- Extract chassis tokens ONLY (e.g. VZJ9#, TRJ12#, GRS20#)
- REMOVE any trailing numeric noise (e.g. 3400, 2700, 2.7, 4.0)

- If token matches chassis pattern → freeze it (DO NOT modify)

---------------------
FILTER
---------------------
Must match:
^[A-Z]{2,5}[0-9]{2,3}[A-Z]?$
^[A-Z]{2,5}[0-9]{2}#$

---------------------
OUTPUT
---------------------
{"tecdoc_types":["ACR50"]}

---------------------
EXAMPLES
---------------------

1ZR/2ZR/3ZR
→ {"tecdoc_types":["ZRE15#","ZRE18#","ZRE17#"]}

2GR/IS/GS
→ {"tecdoc_types":["GSE2#","GRS19#"]}

VZJ9#3400
→ {"tecdoc_types":["VZJ9#"]}

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

	// ✅ Normalize + filter
	final := make([]string, 0)
	for _, t := range result.TecdocTypes {
		t = strings.ToUpper(strings.TrimSpace(t))

		if t == "" {
			continue
		}

		// ❌ remove platform codes like XV70
		if strings.HasPrefix(t, "XV") {
			continue
		}

		// ✅ basic TecDoc pattern (letters+numbers)
		if len(t) >= 5 {
			final = append(final, t)
		}
	}

	if len(final) > 0 {
		return final, nil
	}

	return result.TecdocTypes, nil
}
