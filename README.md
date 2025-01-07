# go-osprofiles

A Go library to simplify the creation and management of application profiles native to the OS

# Supported storage drivers

1. OS keyring
2. in-memory
3. encrypted on the file system

# Next steps

This project was born out of [OpenTDF](https://github.com/opentdf/platform) and [otdfctl](https://github.com/opentdf/otdfctl).

Further next steps:

1. update store abstraction to more of a typical key/value?
2. make the value stored for each profile more generic (`interface{}` or `map[string]interface{}`)
3. tests / CI
4. docs
5. example CLI
