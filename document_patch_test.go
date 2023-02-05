package nimona

// func TestDocumentPatch_ApplyDocumentPatch(t *testing.T) {
// 	a := &CborFixture{
// 		String: "foo",
// 	}

// 	b := &CborFixture{
// 		String: "bar",
// 		Int64:  42,
// 	}

// 	aCbor, err := MarshalCBORBytes(a)
// 	require.NoError(t, err)

// 	bCbor, err := MarshalCBORBytes(b)
// 	require.NoError(t, err)

// 	p, err := CreateDocumentPatch(aCbor, bCbor)
// 	require.NoError(t, err)

// 	err = ApplyDocumentPatch(a, p)
// 	require.NoError(t, err)
// 	require.Equal(t, b, a)
// }

// func TestDocumentPatch_CreateDocumentPatch(t *testing.T) {
// 	a := &CborFixture{
// 		String: "foo",
// 	}

// 	b := &CborFixture{
// 		String: "bar",
// 		Int64:  42,
// 	}

// 	aCbor, err := MarshalCBORBytes(a)
// 	require.NoError(t, err)

// 	bCbor, err := MarshalCBORBytes(b)
// 	require.NoError(t, err)

// 	p, err := CreateDocumentPatch(aCbor, bCbor)
// 	require.NoError(t, err)

// 	id := NewTestIdentity(t)
// 	p.Metadata.Owner = id

// 	p.Dependencies = []DocumentID{{
// 		DocumentHash: NewRandomHash(t),
// 	}}
// }

// func TestDocumentGraph(t *testing.T) {
// 	a := &CborFixture{
// 		String: "foo",
// 	}

// 	// b := &CborFixture{
// 	// 	String: "bar",
// 	// 	Int64:  42,
// 	// }

// 	rootDoc, err := NewDocumentMap(a)
// 	require.NoError(t, err)

// 	g, err := NewDocumentGraph(rootDoc)
// 	require.NoError(t, err)

// 	g.CreatePatch(rootDoc)
// }
