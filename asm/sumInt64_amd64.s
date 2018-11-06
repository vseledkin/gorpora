// func SumInt64(xs []int64) int64
TEXT ·SumInt64(SB),7,$0
    MOVQ    $0, SI       // n
    MOVQ    xs_data+0(FP), BX // BX = &xs[0]
    MOVQ    xs_len+8(FP), CX // len(xs)
start:
    DECQ    CX           // CX--
    JL done              // jump if CX = 0
    ADDQ    (BX), SI     // n += *BX
    ADDQ    $8, BX       // BX += 8
    JMP start

done:
    MOVQ    SI, r+24(FP) // return n
    RET
