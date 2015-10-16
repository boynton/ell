gell
====

A Go implementation of the [Ell](https://github.com/boynton/ell) language

## Features

All basic data types implemented (null, boolean, string, character, number, symbol, keyword, type, list, vector, struct,
instance, blob, function, error, channel).

The reader handles all of [EllDN](https://github.com/boynton/elldn) syntax, which is a superset of JSON.

Function binding includes option and keyword arguments, as well as rest arguments. See tests/argbinding_test.ell for
examples of that.

Simple macros are supported, see tests/macro_example.ell and lib/ell.ell for examples of those.

Full continuations are supported. See tests/continuation_test.ell for examples.

Typed objects, called `instances` are also supported. An instance is simply a pairing of type and value.
They are used by defstruct and deftype, which are defined as macros
in lib/ell.ell, with some examples also that file, as well as tests/defstruct_examples.ell.

Generic functions with full multimethod dispatch is also supported, defined in lib/ell.ell, with an example in
tests/multimethod_examples.ell.

Keywords are actually symbols that have trailing colons, so they can be used anywhere symbols are used, except
the compiler knows they evaluate to themselves. They can be treated as accessor functions on maps, also.

Similarly, types are  symbols surround by angle brackets, and they are also self-evaluating.

The various files in the tests/ directory contain a variety of examples

## Usage

Both a REPL and execution of a file is supported:

    $ go get github.com/boynton/gell
    $ gell
    gell v0.2
    ? 5
    = 5
    ? +
    = #[function +]
    ? (+ 2 3)
    = 5
    ? <ctrl-d>
    $ gell $GOPATH/src/github.com/boynton/gell/tests/tests.ell
    [all tests passed]
    $
