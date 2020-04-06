API calls:
- send api should:
  - take a topic directly, `psstopic` should be a type, with its own serialization
  - take a string or byte slice (this should be the payload) 
  - list of byte slices (list of targets, which are short sequences of bits which are matched after the hash) 
    - (`Address` type can be re-used here)

- put: 
  - if it needs a chunk, the chunk is constructed from the chunker structure
  - 

  `NewHasherStore`
  `storeChunk`


in between the api and the store call:
- init trojan message construct (topic, payload), calculates length, padding, serializes
- iterator (init with the span, target prefixes) → try the nonce, output for now the chunk itself 
- store.put() (mode put upload)

we should validate that lengths are ok, for this we can define the types as fixed-size byte arrays rather than slices