//func Dxmulelyplusz(X, Y []float64)
TEXT ·Dxmulelyplusz(SB), 7, $0
	MOVQ	X_data+0(FP), SI
	MOVQ	X_len+8(FP), BP
	MOVQ	Y_data+24(FP), CX
	MOVQ	Z_data+48(FP), DI

	SUBQ	$2, BP
	JL		rest	// There are less than 2 pairs to process
	simd_loop:
		// Load four pairs and scale
		MOVUPD	(SI), X2
		MOVUPD	(CX), X3
		MULPD	X3, X2
		MOVUPD	(DI), X1
		ADDPD	X2, X1
		MOVUPD	X1, (DI)

		// Update data pointers
		ADDQ	$16, SI
		ADDQ	$16, CX
		ADDQ	$16, DI

		SUBQ	$2, BP
		JGE		simd_loop	// There are 2 or more pairs to process
	JMP	rest

rest:
	// Undo last SUBQ
	ADDQ	$2,	BP
	// Check that are there any value to process
	JE	end
	loop:
		MOVSD	(SI), X2
		MULSD	(CX), X2
		ADDSD	(DI), X2
		MOVSD	X2, (DI)

		// Update data pointers
		ADDQ	$2, SI
		ADDQ	$2, CX
		ADDQ	$2, DI

		DECQ	BP
		JNE	loop
	RET

end:
	RET
