## gorpora

Random set of utilities for text corpora processing, primarily for machine learning purpoces.

##### Usage of gorpora:

## gorpora command [arguments]

commands are:

## fb2text


Parameters:

-  -i string
    	directory with fb2 files, will be processed recursively
-  -l int
    	number of \n's added after each text output block (paragraph) (default 1)
-  -t int
    	number of threads for parallel processing of conversion jobs (default 1)
    	
## normalize.html.entities

Parameters:

-  -debug
    	do othing only print use cases
-  -max int
    	maximum number of lines to process
    	
## strip.html


## word.tokenizer

Parameters:

-  -debug
    	do nothing only print use cases
-  -lemma
    	output lemmas instead of words
-  -udpipe
    	use Udpipe as tokenizer
    	
## sentence.tokenizer

Parameters:

-  -debug
    	do nothing only print use cases
-  -max int
    	maximum sentence length in chars (default 1000000)
-  -min int
    	minimun sentence length in chars (default 10)
    	
## filter.language

Parameters:

-  -debug
    	do othing only print use cases
-  -lang value
    	set of accepted languages

## unique
 
accepts text lines to stdin, outputs to stdout filtering out non unique lines. 

Parameters: 

- -debug
    do nothing only print use cases
