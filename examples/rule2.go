package main

import (
	"encoding/json"
	"fmt"

	"github.com/oarkflow/expr"

	"github.com/oarkflow/filters"
	"github.com/oarkflow/filters/utils"
)

var requestData = []byte(`
{
    "patient_header": {
        "disch_disp": null,
        "transfer_dest": null,
        "patient_status": null,
        "admit_date": null,
        "injury_date": null,
        "lmp_date": null,
        "status": "IN_PROGRESS",
        "message": null,
        "patient_dob": "2019-05-02",
        "patient_sex": "M"
    },
    "coding": [
        {
            "dos": "2020/01/01",
            "details": {
                "pro": {
                    "em": {
                        "em_modifier1": "8",
                        "em_downcode": false,
                        "shared": false
                    },
                    "downcode": [],
                    "special": [],
                    "cpt": [
                        {
                            "procedure_num": "AN65450",
                            "procedure_qty": 1,
                            "billing_provider": null,
                            "secondary_provider": null
                        },
                        {
                            "procedure_num": "AN65450",
                            "procedure_qty": 1,
                            "billing_provider": null,
                            "secondary_provider": null
                        }
                    ],
                    "hcpcs": []
                },
                "fac": {
                    "em": null,
                    "special": [],
                    "cpt": [],
                    "hcpcs": []
                },
                "dx": {
                    "pro": [],
                    "fac": []
                },
                "cdi": {
                    "pro": [],
                    "fac": []
                },
                "notes": []
            }
        }
    ]
}
`)

func main() {
	expr.AddFunction("age", utils.BuiltinAge)

	var data map[string]any
	err := json.Unmarshal(requestData, &data)
	if err != nil {
		panic(err)
	}

	filter := filters.NewFilter("coding.#.details.pro.cpt.#.procedure_num", filters.GreaterThanCount, 2)
	lookup := &filters.Lookup{
		Data: []map[string]any{
			{
				"code":         "AN65450",
				"charge_type":  "ED_PROFEE",
				"work_item_id": 33,
			},
		},
		Condition: "map(filter(lookup, .charge_type == 'ED_PROFEE'), .code)",
	}
	filter.SetLookup(lookup)
	fmt.Println(filters.Match(data, filter))
}
