gell
====

A Go implementation of the [Ell](https://github.com/boynton/ell) language

## Features

All basic data types implemented (null, boolean, string, number, symbol, keyword, type, list, array, struct, object, function).

The reader handles all of [EllDN](https://github.com/boynton/elldn) syntax, which is a superset of JSON.

The lambda binding includes option and keyword arguments, as well as rest arguments. See tests/argbinding_test.ell for
examples of that.

Simple macros are supported, see tests/macro_example.ell and lib/ell.ell for examples of those.

Type objects, called `instances` are also supported. They are used by defstruct and deftype, define as macros
in lib/ell.ell, with some examples also that file, as well as tests/defstruct_examples.ell.

Generic functions with full multimethod dispatch is also supported, defined in lib/ell.ell, with an example in
tests/multimethod_examples.ell.

Keywords are actually symbols that have trailing colons, so they can be used anywhere symbols are used, except
the compiler knows they evaluate to themselves. They can be treated as accessor functions on maps, also.

Similarly, types are  symbols surround by angle brackets, and they are also self-evaluating.

## Usage

Both a REPL and execution of a file is supported:

    $ go get github.com/boynton/gell
    $ gell
    gell v0.1
    ? 5
    = 5
    ? +
    = #[function + (<number>*) <number>]
    ? (+ 2 3)
    = 5
    ? <ctrl-d>
    $ gell tests/tests.ell
    [all tests passed]
    $
