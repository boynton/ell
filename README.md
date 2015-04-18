gell
====

A Go implementation of the Ell language

## Features

All basic data types implemented (null, boolean, character, string, number, list, vector, map, function).

The reader handles most of EllDN, including map literals.

The lambda binding includes option and keyword arguments, as well as rest arguments. See tests/argbinding_test.ell for
examples of that.

Keywords are actually symbols that have leading or trailing colons, so they can be used anywhere symbols are used, except
the compiler knows they evaluate to themselves. They can be treated as accessor functions on maps.

## Usage

Both a REPL and execution of a file is supported:

    $ go get github.com/boynton/gell
    $ gell
    ? 5
    = 5
    ? +
    = <function +>
    ? (+ 2 3)
    = 5
    ? <ctrl-d>
    $ gell tests/argbinding_test.ell
    [all tests passed]
    $

