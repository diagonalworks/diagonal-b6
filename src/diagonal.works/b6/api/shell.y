%{

package api

import (
    "diagonal.works/b6"
)

%}

%token ',' '(' ')' '|' '{' '}' '[' ']' '=' '&' ':'
%token <e> FLOAT
%token <e> INT
%token <e> FEATURE_ID
%token <e> SYMBOL
%token <e> STRING
%token <e> TAG_KEY
%token <e> ARROW

%union {
	e b6.Expression
    es []b6.Expression
}

%type <e> expression latlng tag call arg pipeline lambda collection collection_items collection_key_value collection_key collection_value group tagvalue query query_expression query_tag
%type <es> args symbols

%%

top:
    pipeline
    {
        yylex.(*lexer).Top = reduceRootCall($1, yylex.(*lexer))
    }

pipeline:
    pipeline '|' call
    {
        $$ = reducePipeline($1, $3, yylex.(*lexer))
    }
|   call

expression:
    latlng
|   tag
|   lambda
|   collection
|   group
|   query
|   STRING
|   FLOAT
|   INT
|   FEATURE_ID

latlng:
    FLOAT ',' FLOAT
    {
        $$ = reduceLatLng($1, $3, yylex.(*lexer))
    }

tag:
    TAG_KEY '=' tagvalue
    {
        $$ = reduceTag($1, $3, yylex.(*lexer))
    }
|   SYMBOL '=' tagvalue
    {
        $$ = reduceTag($1, $3, yylex.(*lexer))
    }

call:
    SYMBOL
    {
        $$ = reduceCall($1, yylex.(*lexer))        
    }
|   SYMBOL args
    {
        $$ = reduceCallWithArgs($1, $2, yylex.(*lexer))
    }
|   expression

args:
    args arg
    {
        $$ = reduceArgs($1, $2)
    }
|   arg
    {
        $$ = reduceArg($1)
    }

arg:
    SYMBOL
|   expression

lambda:
    '{' symbols ARROW pipeline '}'
    {
        $$ = reduceLambda($2, $4)
    }
|   '{' ARROW pipeline '}'
    {
        $$ = reduceLambdaWithoutArgs($3)
    }

symbols:
    SYMBOL
    {
        $$ = reduceSymbolsSymbol($1)
    }
|   symbols ',' SYMBOL
    {
        $$ = reduceSymbolsSymbols($1, $3)
    }

collection:
    '{' collection_items '}'
    {
        $$ = reduceCollectionItems($2)
    }

collection_items:
    collection_key_value
    {
        $$ = reduceCollectionItemsKeyValue($1)
    }
|   collection_items ',' collection_key_value
    {
        $$ = reduceCollectionItemsItemsKeyValue($1, $3)
    }

collection_key_value:
    collection_key ':' collection_value
    {
        $$ = reduceCollectionKeyValue($1, $3)
    }
|   collection_value
    {
        $$ = reduceCollectionValueWithImplictKey($1)
    }

collection_key:
    STRING
|   INT
|   FEATURE_ID
|   tag
|   group

collection_value:
    STRING
|   INT
|   FEATURE_ID
|   FLOAT
|   tag
|   group

group:
   '(' pipeline ')'
   {
        $$ = $2
   }

query:
    '[' query_expression ']'
    {
        $$ = $2
    }

query_expression:
    query_tag
|   query_tag '&' query_expression
    {
        $$ = reduceAnd($1, $3)
    }
|   query '&' query_expression
    {
        $$ = reduceAnd($1, $3)
    }
|   query_tag '|' query_expression
    {
        $$ = reduceOr($1, $3)
    }
|   query '|' query_expression
    {
        $$ = reduceOr($1, $3)
    }
|   query

query_tag:
    TAG_KEY
    {
        $$ = reduceTagKey($1);
    }
|   TAG_KEY '=' tagvalue
    {
        $$ = reduceTagKeyValue($1, $3);
    }
|   SYMBOL
    {
        $$ = reduceTagKey($1);
    }
|   SYMBOL '=' tagvalue
    {
        $$ = reduceTagKeyValue($1, $3);
    }

tagvalue:
    SYMBOL
|   STRING
