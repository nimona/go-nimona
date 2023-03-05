package nimona

// func TestGetShorthand(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		cborer  Documenter
// 		want    string
// 		wantErr bool
// 	}{{
// 		name: "network handle",
// 		cborer: &NetworkAlias{
// 			Hostname: "nimona.io",
// 		},
// 		want: "core/network/alias",
// 	}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cborBytes, err := MarshalJSON(tt.cborer)
// 			require.NoError(t, err)

// 			got, err := GetDocumentTypeFromCbor(cborBytes)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetShorthand() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			require.Equal(t, tt.want, got)
// 		})
// 	}
// }
