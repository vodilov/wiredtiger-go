# wiredtiger-go
WiredTiger for Go Lang

Under development

##Ready:
- [x] Open connection
- [x] WT_CONNECTION
- [x] WT_SESSION
- [x] Pack/UnPack
- [ ] **WT_CURSOR - TESTING**
- [ ] Documentation
- [ ] WT_CONFIG_PARSER / WT_CONFIG_ITEM
- [ ] WT_COLLATOR
- [ ] WT_EXTRACTOR
- [ ] WT_COMPRESSOR
- [ ] WT_ENCRYPTOR
- [ ] WT_DATA_SOURCE
 

##Format types

| Format | C Type | Java type| Python type | Go Lang Type | Notes |
| --- | --- | --- | --- | --- | --- |
| x | N/A | N/A | N/A | N/A | pad byte, no associated value |
| b | int8_t | byte | int | int8 | signed byte |
| B | uint8_t | byte | int | uint8 | unsigned byte |
| h | int16_t | short | int | int16 | signed 16-bit |
| H | uint16_t | short | int | uint16 | unsigned 16-bit|
| i | int32_t | int | int | int32 | signed 32-bit | 
| I | uint32_t | int | int | uint32 | unsigned 32-bit |
| l | int32_t | int | int | int32 | signed 32-bit |
| L | uint32_t | int | int | uint32 | unsigned 32-bit |
| q | int64_t | long | int | int64 | signed 64-bit |
| Q | uint64_t | long | int | uint64 | unsigned 64-bit |
| r | uint64_t | long | int | uint64 | record number |
| s | char[] | String | str | string | fixed-length string |
| S | char[] | String | str | string | NUL-terminated string |
| t | uint8_t | byte | int | byte | fixed-length bit field |
| u | WT_ITEM * | byte[] | str | []byte | raw byte array |




