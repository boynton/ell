ell
====

ℒ (ell) is a LISP dialect, with semantics similar to Scheme, but adds Keywords, Structs, user-defined Types, and other features.

## Getting started

The reference version of Ell is written in Go. If you have Go installed on your machine, you can install Ell easily:

	go get github.com/boynton/gell

This installs the self-contained binary into `$GOPATH/bin/gell`. Ell loads its library files from locations defined by the `ELL_PATH`
environment variable. If that variable is not defined, the default path is `".:$HOME/lib/ell:$GOPATH/src/github.com/boynton/gell/lib"`.

If you have a `.ell` file in your home directory, it will get loaded and executed when running ell interactively.

	$ gell
	[loading /Users/lee/.ell]
	gell v0.2
	?

The `?` prompt is the Read-Eval-Print-Loop (REPL) for ell, waiting for your input. Entering CTRL-D will end the REPL and exit
back to the shell.

## Primitive types

Ell defines a variety of native data types, all of which have an external textual representation. This data notation is called
EllDN, and is a superset of JSON.

	? 5.2
	= 5.2
	? "five"
	= "five"
	? true
	= true
	? null
	= null
	? [1, 2, 3]
	= [1 2 3]
	? {"x": 1, "y": 2}
	= {"x" 1 "y" 2}

The basic JSON types are supported as you would expect. Although JSON-compatible format is accepted for vectors
(JSON arrays) and structs (JSON objects), the canonical form for both does not include separating commas.

EllDN also introduces _keywords_, _types_, _symbols_, and _lists_ to the syntax.

Keywords are symbolic identifiers that end in a colon (':'), and types are symbolic identifiers surrounded by angle brackets ('<' and '>').
Both are self-evaluating, but used for different purposes:

	? foo:
	= foo:
	? <string>
	= <string>

Symbols are identifiers that form variable references in Ell, so are interpreted differently. Lists are ordered
sequences, and form the syntax for function (or macro) calls in Ell:

	? length				; a reference to the `length` variable
	= #[function length]
	? (length "foo")        ; a function call to the length function
	= 3

Most types evaluate to themselves. Since symbols and lists do not evaluate to themselves, they must be _quoted_ to be taken
literally:

	? x
	 *** [error: Undefined symbol: x]
	? 'x
	= x
	? (f 23)
 	 *** [error: Undefined symbol: f]
	? '(f 23)
	= (f 23)

The vector and struct elements are also evaluated, so you sometimes need to quote them, too (as that is more convenient
than quoting every interior item):

	? [1 two 3]
	 *** [error: Undefined symbol: two] 
	? [1 'two 3]      ; you can quote jsut what needs to be quoted
	= [1 two 3]
	? '[1 two 3]      ; or jsut quote the whole thing
	= [1 two 3]
	? {x 2}
	 *** [error: Undefined symbol: x]
	? '{x 2}
	= {x 2}
	? {"x" two}
	 *** [error: Undefined symbol: two]
	? '{"x" two}
	= {"x" two}

The complete list of primitive types in Ell is:

	* <null>
	* <boolean>
	* <character>
	* <string>
	* <blob>
	* <symbol>
	* <keyword>
	* <type>
	* <list>
	* <vector>
	* <struct>
	* <function>
	* <code>
	* <error>
	* <channel>

You can define additional types in terms of other types, this is discussed later.

## Core expressions

	* <symbol> - variable reference
	* (quote expr) - literal data
	* (do expr1 expr2 ...) - expression sequencing
	* (if pred consequent antecedent) - conditional
	* (<function> <any> ...) - function call
	* (fn (arg1 arg2 ...) expr ...) - function creation
	* (set! name expr) - sets the lexically apparent variable to the value
	* (def name expr) - define value. At the top level, sets the global variable. Inside a function, creates a new frame with the binding.
	* (defmacro name args expr ...) - define a new macro

### Variables and literals

As mentioned before, most data items evaluate to themselves. But symbols and lists do not. When a symbol
is interpreted, the closest lexical binding of that variable is looked up, starting from the innermost
frame, and ending with the global environment.

	? +
	= #[function +]  		; the global binding for + is the addition function
	? (defn f (x) (+ x 1))	; the x here is defined only inside the function
	= #[function f]
	
The `quote` primitive is used to prevent the normal evaluation if its argument. It is the long form of
the single quote reader macro, which actually just produces a quote form:

	? (quote foo)
	= foo
	? 'foo ; a shorthand for the same thing
	= foo
	? (def x 23)
	= 23
	? x
	= 23
	? 'x
	= x

### Conditionals and sequencing

The primitive for conditionals is `if`, which takes a predicate, and if the predicate is true evaluates the
consequent. An optional antecedent clause will be executed of the predicate is false.

	? (if true 'yes)
	= yes
	? (if false 'yes)
	= null
	? (if false 'yes 'no)
	= no

Sometimes more than one expression is needed in the place of one, largely for side-effects. This is what the `do` special
form is for:

	? (do (println "hello") 'blah)
	hello
	= blah
	? (if true (do (println "it was true!") 1) (do (println "it wasn't true!") 0))
	it was true!
	= 1


### Functions and lexical environments

A list is interpreted as a function call. For example, the following applies the `+` function to the arguments 2 and 3:

	? (+ 2 3)
	= 5

Ell provides a variety of primitive functions (like `+`). New functions are defined by the `fn` special form:

	? (fn (x) (+ 1 x))
	= #[function]

This creates an anonymous function that takes a single argument `x` and returns the sum of that and 1. This
actually creates a closure over a new frame in the environment containing the variable `x`, and executes the body of the function
in that lexical environment. The function returned is a first class object that itself can be passed around.
All other binding forms can be defined in terms of this primitive, for example, consider the following:

	(let ((x 23)) (+ 1 x))

The `let` special form is actually just a macro that generates the equivalent primitive form:

	? (macroexpand '(let ((x 23)) (+ 1 x)))
	= ((fn (x) (+ 1 x)) 23)

A function lives on with indefinite extent, closed over any variables in its lexical environment. For example:

	? (def f (let ((counter 0)) (fn () (set! counter (inc counter)) counter)))
	= #[function f]
	? (f)
	= 1
	? (f)
	= 2

In the above example, the primitive expression `set!` is also shown. It sets the value for the variable
determined by its first argument (a symbol).

Normally, functions are defined at the top level, i.e. in the global environment, using the `defn` special form
(which itself is just a macro):

	? (defn f (x) (+ 1 x))
	= #[function f]
	? (f 23)
	= 24

If def and defn are used inside a function, they create a new binding inside the function, rather than side-effect
the global values for the symbols.

The `defmacro` primitive form allows the definition of new special forms (i.e. syntactic constructs used by the compiler).

	? (defmacro blah (lst x) `(cons ~x ~lst))
	= blah
	? (blah '(1 2) 23)
	= (23 1 2)
	? (macroexpand '(blah '(1 2) 23))
	= (cons 23 '(1 2))

This example shows the `quasiquote` macro, which simulates a simple quote, but allows escaped values to be inserted.
In general `~x` means "insert the current value of x here", and `~@x` means "splice the list represented by x into the
expression here".


#### Function argument binding forms

In addition to traditional lambda definitions, with explicit arguments and/or "rest" arguments, optional named arguments with defaults, and keyword arguments, are also supported:

	? (defn f (x y) (list x y))
	= #[function f]
	? (f 1 2)
	= (1 2)
	? (defn f (x & rest) (list x rest))
	= #[function f]
	? (f 1 2)
	= (1 (2))
	? (f 1 2 3)
	= (1 (2 3))
	? (defn f args args)
	= #[function f]
	? (f 1 2 3)
	= (1 2 3)
	? (defn f (x [y]) (list x y))
	= #[function f]
	? (f 1 2)
	= (1 2)
	? (f 1)
	= (1 null)
	? (defn f (x [(y 23)]) (list x y))
	= #[function f]
	? (f 1)
	= (1 23)
	? (f 1 2)
	= (1 2)
	? (defn f (x {y: 23 z: 57}) (list x y z))
	= #[function f]
	? (f 1)
	= (1 23 57)
	? (f 1 2)                                                                                               
	 *** Bad keyword arguments: [2] 
	? (f 1 y: 2)
	= (1 2 57)
	? (f 1 z: 2)
	= (1 23 2)
	? (f 1 z: 2 y: 3)
	= (1 3 2)

## Defining new types

The `type` function returns the type of its argument:

	? (type 5)
	= <number>
	? (type "foo")
	= <string>
	? (type <string>)
	= <type>
	
Types are referred to symbolically, and also evaluate to themselves. They do not have to be defined to
be referred to, as they are essentially just tags on data. A special syntax allows them to be read/written:

	? <foo>
	= <foo>	
	? (type <foo>)
	= <type>
	? (type #<foo>"blah")
	= <foo>
	? (value #<foo>"blah")
	= "blah"


New types can be introduced by attaching to other data objects. For example:

	? (def x (instance <foo> "blah"))
	= #<foo>"blah"
	? (type x)
	= <foo>
	? (value x)
	= "blah"

Some convenience macros for defining new types are provided:

	? (deftype foo (o) (and (string? o) (< (length o) 5)))
	= <foo>
	? (foo "blah")
	= #<foo>"blah"
	? (foo "no way")                  
 	 *** [syntax-error: not a valid <foo>:  "no way"] [in foo]
	? (foo? (foo "blah"))
	= true
	? (foo? "blah")
	= false

It is defined in terms of a validation predicate. Once defined, the type is independent, and
has no relation to any other type; there is no inheritance of types.

Defining a type that is a struct with certain fields is common enough that a dedicated macro is defined
to make it simpler:

	? (defstruct point x: <number> y: <number>)
	= <point>
	? (point)
	 *** [validation-error: type <point> missing field x: {}] [in point] 
 	? (point x: 1 y: 2)
	= #<point>{x: 1 y: 2}
	? (def data {x: 1 y: 2})
	= {x: 1 y: 2}
	? (struct? data)
	= true
	? (point? data)
	= false
	? (def pt (point data))
	= #<point>{x: 1 y: 2}
	? (struct? pt)
	= false
	? (point? pt)
	= true
	? (value pt)
	= {x: 1 y: 2}
	? (type (value pt))
	= <struct>
	? (equal? data (value pt))
	= true
	? (identical? data (value pt)) ; not identical means a copy was made.
	= false
	? (point-fields)
	= {x: <number> y: <number>}

To access fields of a struct, including any defined as above, the `get` function can be used, but since all fields
have keywords as names, using a keyword as a function is more idiomatic:

	? (x: pt)
	= 1
	? (y: pt)
	= 2

Although to change the values, put! must be used on the value of the instance. Any such mutable operation is not
encouraged, but sometimes necessary:

	? (put! data x: 23)
	= null
	? data
	= {x: 23 y: 2}
	? (put! pt x: 23)
	 *** [argument-error: put! expected a <struct> for argument 1, got a <point>] 
	? (put! (value pt) x: 57)       
	= null
	? pt
	= #<point>{x: 57 y: 2}

## Defining methods on types

Ell provides generic method dispatch, supporting multimethods. This means any number of arguments
to a function may be specialized, and dispatch can be based on all of them.

For example:

	? (defgeneric add (x y))
	= #[function add]
	? (defmethod add ((x <string>) y) (string x "|other|" y))
	= add
	? (defmethod add ((x <number>) (y <number>)) (+ x y))
	= add
	? (defmethod add ((x <string>) (y <string>)) (string x "|string|" y))
	= add
	? (defmethod add ((x <list>) (y <list>)) (concat x y))
	= add
	? (defmethod add ((x <vector>) (y <vector>)) (apply vector (concat (to-list x) (to-list y))))
	= add
	? (add 1 2)
	= 3
	? (add "foo" "bar")
	= "foo|string|bar"
	? (add "foo" 'bar)
	= "foo|other|bar"
	? (add '(1 2) '(3 4))
	= (1 2 3 4)
	? (add [1 2] [3 4])
	= [1 2 3 4]


## Other features

### Continuations


## More info



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
