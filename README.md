[![experimental](http://badges.github.io/stability-badges/dist/experimental.svg)](http://github.com/badges/stability-badges)
[![Go Report Card](https://goreportcard.com/badge/github.com/devblok/koru)](https://goreportcard.com/report/github.com/devblok/koru)
[![Test Coverage](https://api.codeclimate.com/v1/badges/2a8868a10a3e307f1189/test_coverage)](https://codeclimate.com/github/devblok/koru/test_coverage)

# Koru3D

At this stage, this project is merely my playground to try out Vulkan along
with a bunch of different things both in Go and technology in general.

## Things to try (roughly in descending order of priority):
- [x] Vulkan
- [x] [Packr](https://github.com/gobuffalo/packr)
- [x] gotk3 GUI that's loaded from glade resources
- [ ] Try running the thing on Android
- [ ] Memory mapped resource packs
- [ ] [Tengo](https://github.com/d5/tengo) scripting engine
- [ ] A Plugin system
- [ ] Ability to run scripts on all cores
- [ ] Ability to run Vulkan rendering on all cores
- [ ] A nice, sufficient editor on gotk3
- [ ] An AI engine (experiment with neural networks too)
- [ ] Artificial intelligence that utilises GPU? [gorgonia](https://github.com/gorgonia/gorgonia)
- [ ] Design an actual game

## Known bugs

- go-sdl does not yet release vulkan bindings, therefore Gopkg contains a commit from master. `dep ensure` likes to reset the verision to a newest tag. Keep an eye on it when updating packages.