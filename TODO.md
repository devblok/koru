# TODO list and achievables

Small feats, that consist of a list of TODO items, to consider them complete

### Triangle

- [x] Not included already complete tasks...
- [x] Correct window resizing (recreation of Renderer properties)
- [x] Triangle is a model, not a hard-coded shader
- [x] ~~Triangle on korued window~~ Don't want to deal with gotk3 right now, it hasn't got VulkanContext exposed to Go (contribute at later time)
- [x] Ability to disable frames per second limitation
- [x] Operation in 3D space, not just screen 2D plane

### Model meshes

- [ ] Mesh loaded from Collada format
- [ ] Gob encoder (accessible through korucli)
- [ ] Gob model loader
- [ ] Basic histogram capability (for later benchmarks and compiler experimentation)
- [ ] Textures

## Good hygene

- [x] Strings given to vulkan need to be escaped
- [x] Command pools should be freed, not re-allocated
- [ ] Multiple sets of resources per frame + fences
- [ ] Contribute by adding VulkanContext to gotk3 

## Editor

- [ ] Logs piped to korued console

## Support

- [ ] MacOS build
- [ ] Windows build
- [ ] Travis pipeline