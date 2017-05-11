## v3

- gopkg.in is not supported any more. Update import path to github.com/vmihailenco/msgpack.
- Msgpack maps are decoded into map[string]interface{}.
- EncodeSliceLen is removed in favor of EncodeArrayLen. DecodeSliceLen is removed in favor of DecodeArrayLen.
