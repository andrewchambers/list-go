# list-go

This package is fork of the go standard library container/list.

Differences are summarised below:

- Support generics.
- Each List has a small pool of removed elements for reuse.
  - You cannot use *Element after it is removed from a list