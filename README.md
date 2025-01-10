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
2. tests
3. docs
4. test OS platform directories work as desired (see ./pkg/platform and various TODO comments)
