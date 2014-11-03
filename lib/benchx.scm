(define sort_sorted? (lambda (seq less?) (if (null? seq) #t (if (vector? seq) ((lambda (n) (if (<= n 1) #t ((lambda (system_loop) (set! system_loop (lambda (i) (if ((lambda (system_tmp) (if system_tmp system_tmp (less? (vector-ref seq (- i 1)) (vector-ref seq i)))) (= i n)) (= i n) (system_loop (+ i 1))))) (system_loop 1)) #f))) (vector-length seq)) ((lambda (loop) (set! loop (lambda (last next) ((lambda (system_tmp) (if system_tmp system_tmp (if (not (less? (car next) last)) (loop (car next) (cdr next)) #f))) (null? next)))) (loop (car seq) (cdr seq))) #f)))))
(define sort_merge (lambda (a b less?) (if (null? a) b (if (null? b) a ((lambda (loop) (set! loop (lambda (x a y b) (if (less? y x) (if (null? b) (cons y (cons x a)) (cons y (loop x a (car b) (cdr b)))) (if (null? a) (cons x (cons y b)) (cons x (loop (car a) (cdr a) y b)))))) (loop (car a) (cdr a) (car b) (cdr b))) #f)))))
(define sort_merge! (lambda (a b less?) ((lambda (loop) (set! loop (lambda (r a b) (if (less? (car b) (car a)) (begin (set-cdr! r b) (if (null? (cdr b)) (set-cdr! b a) (loop b a (cdr b)))) (begin (set-cdr! r a) (if (null? (cdr a)) (set-cdr! a b) (loop a (cdr a) b)))))) (if (null? a) b (if (null? b) a (if (less? (car b) (car a)) (begin (if (null? (cdr b)) (set-cdr! b a) (loop b a (cdr b))) b) (begin (if (null? (cdr a)) (set-cdr! a b) (loop a (cdr a) b)) a))))) #f)))
(define sort_sort! (lambda (seq less?) ((lambda (step) (set! step (lambda (n) (if (> n 2) ((lambda (j) ((lambda (a) ((lambda (k) ((lambda (b) (sort_merge! a b less?)) (step k))) (- n j))) (step j))) (quotient n 2)) (if (= n 2) ((lambda (x y p) (set! seq (cdr (cdr seq))) (if (less? y x) (begin (set-car! p y) (set-car! (cdr p) x)) #f) (set-cdr! (cdr p) '()) p) (car seq) (car (cdr seq)) seq) (if (= n 1) ((lambda (p) (set! seq (cdr seq)) (set-cdr! p '()) p) seq) '()))))) (if (vector? seq) ((lambda (n) (set! seq (vector->list seq)) ((lambda (system_loop) (set! system_loop (lambda (i p) (if (null? p) vector (begin (vector-set! vector i (car p)) (system_loop (+ i 1) (cdr p)))))) (system_loop 0 (step n))) #f)) (vector-length seq)) (step (length seq)))) #f)))
(define sort_sort (lambda (seq less?) (if (vector? seq) (list->vector (sort_sort! (vector->list seq) less?)) (sort_sort! (append seq '()) less?))))
(define sorted? sort_sorted?)
(define merge sort_merge)
(define merge! sort_merge!)
(define sort sort_sort)
(define sort! sort_sort!)

(define pi (lambda (n d) (set! n (+ (quotient n d) 1)) ((lambda (m) ((lambda (r a) (vector-set! a m 4) ((lambda (system_loop) (set! system_loop (lambda (b q j) (if (> j n) #f (begin (begin ((lambda (system_loop) (set! system_loop (lambda (k) (if (zero? k) #f (begin (begin (set! q (+ q (* (vector-ref a k) r))) ((lambda (t) (vector-set! a k (remainder q t)) (set! q (* k (quotient q t)))) (+ 1 (* 2 k)))) (system_loop (- k 1)))))) (system_loop m)) #f) ((lambda (s) ((lambda (system_loop) (set! system_loop (lambda (l) (if (>= l d) (display s) (begin (display #\0) (system_loop (+ 1 l)))))) (system_loop (string-length s))) #f)) (number->string (+ b (quotient q r)))) (display (if (zero? (modulo j 10)) #\newline #\space))) (system_loop (remainder q r) 0 (+ 1 j)))))) (system_loop 2 0 1)) #f) (newline)) ((lambda (system_loop) (set! system_loop (lambda (i s) (if (>= i d) s (system_loop (+ 1 i) (* 10 s))))) (system_loop 0 1)) #f) (make-vector (+ 1 m) 2))) (quotient (* n d 3322) 1000))))
(define make-foo (lambda (n) ((lambda (system_loop) (set! system_loop (lambda (l i) (if (>= i n) l (system_loop (cons i l) (+ i 1))))) (system_loop '() 0)) #f)))
(define benchmark (lambda () ((lambda (foo) (sort foo <) (sort foo >) (pi 1000 5)) (make-foo 100000))))

((lambda (iterations) ((lambda (loop) (set! loop (lambda (i) (if (> i 0) (begin (benchmark) (loop (- i 1))) #f))) (loop iterations)) #f)) 5)