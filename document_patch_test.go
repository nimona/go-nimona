package nimona

// func TestDocumentPatch_CreateAndApply(t *testing.T) {
// 	original := &CborFixture{
// 		String: "foo",
// 	}

// 	target := &CborFixture{
// 		String: "bar",
// 		Int64:  42,
// 	}

// 	patch, err := CreateDocumentPatch(
// 		original.Document(),
// 		target.Document(),
// 	)
// 	require.NoError(t, err)

// 	applied, err := ApplyDocumentPatch(
// 		original.Document(),
// 		patch,
// 	)
// 	require.NoError(t, err)
// 	require.Equal(t, target, applied)
// }
