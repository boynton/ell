
(define pi
    (lambda (n d)
        (set! n (+ (quotient n d) 1))
        ((lambda (m)
             ((lambda (r a)
                  (vector-set! a m 4)
                  ((lambda (system_loop)
                       (set! system_loop
                           (lambda (b q j)
                               (if (> j n)
                                   #f
                                   (begin
                                       (begin
                                           ((lambda (system_loop)
                                                (set! system_loop
                                                    (lambda (k)
                                                        (if (zero? k)
                                                            #f
                                                            (begin
                                                                (begin
                                                                    (set! q
                                                                        (+ q
                                                                           (* (vector-ref
                                                                                  a
                                                                                  k)
                                                                              r)))
                                                                    ((lambda (t)
                                                                         (vector-set!
                                                                             a
                                                                             k
                                                                             (remainder
                                                                                 q
                                                                                 t))
                                                                         (set! q
                                                                             (* k
                                                                                (quotient
                                                                                    q
                                                                                    t))))
                                                                     (+ 1
                                                                        (* 2
                                                                           k))))
                                                                (system_loop
                                                                    (- k
                                                                       1))))))
                                                (system_loop m))
                                            #f)
                                           ((lambda (s)
                                                ((lambda (system_loop)
                                                     (set! system_loop
                                                         (lambda (l)
                                                             (if (>= l d)
                                                                 (display s)
                                                                 (begin
                                                                     (display
                                                                         "0")
                                                                     (system_loop
                                                                         (+ 1
                                                                            l))))))
                                                     (system_loop
                                                         (string-length s)))
                                                 #f))
                                            (number->string
                                                (+ b (quotient q r))))
                                           (display
                                               (if (zero? (modulo j 10))
                                                   "\n"
                                                   " ")))
                                       (system_loop
                                           (remainder q r)
                                           0
                                           (+ 1 j))))))
                       (system_loop 2 0 1))
                   #f)
                  (newline))
              ((lambda (system_loop)
                   (set! system_loop
                       (lambda (i s)
                           (if (>= i d) s (system_loop (+ 1 i) (* 10 s)))))
                   (system_loop 0 1))
               #f)
              (make-vector (+ 1 m) 2)))
         (quotient (* n (* d 3322)) 1000))))

(pi 1000 5)
