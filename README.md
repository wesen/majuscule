# Majuscule

HashTag completion service for accessibility use.

## TODO

### Functionality

#### Algorithm

- [x] iterative (not recursive) dynamic programming algorithm
- [/] heuristics driven dynamic programming
- [/] word frequency driven heuristic (not just length)

#### Misc 

- [x] vaporwave test frontend
- [ ] statistics
- [ ] hashtag suggestion through search 

#### Indexing new words

- [ ] storage backend
- [ ] indexing new words
- [ ] endpoint for hashtag selection (to update index)

### Connectivity

- [/] grpcweb bindings
- [ ] websockets bindings
- [/] openapi bindings

### Deployment

- [ ] scaffolding for AWS, digitalocean and self-hosting
- [ ] rate limiting
- [ ] performance monitoring
- [ ] logging
- [ ] security (sessions?)
- [ ] benchmarking

### Brainstorm

- [ ] extend to per user suggestion
- [ ] extend to user handle suggestion
