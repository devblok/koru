# TODO list and achievables

Small feats, that consist of a list of TODO items, to consider them complete

### Model meshes

- [x] Mesh loaded from Collada format
- [ ] Gob encoder (accessible through korucli)
- [ ] Gob model loader
- [ ] Basic histogram capability (for later benchmarks and compiler experimentation)
- [x] Textures

## Good hygene

- [x] Strings given to vulkan need to be escaped
- [x] Command pools should be freed, not re-allocated
- [x] Multiple sets of resources per frame + fences
- [ ] Contribute by adding VulkanContext to gotk3
- [ ] Find a way to intelligently set the number of descriptor sets to account for in a pool

## Editor

- [ ] Logs piped to korued console

## Support

- [ ] MacOS build
- [ ] Windows build
- [ ] Travis pipeline