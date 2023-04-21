%{

package api

import (
    pb "diagonal.works/b6/proto"
)

%}

%token ',' '(' ')' '|' '{' '}' '[' ']' '=' '&'
%token <node> FLOAT
%token <node> INT
%token <node> FEATURE_ID
%token <node> SYMBOL
%token <node> STRING
%token <node> TAG_KEY
%token <node> ARROW

%union {
	node *pb.NodeProto
    nodes []*pb.NodeProto
}

%type <node> expression latlng tag call arg pipeline lambda group tagvalue query query_expression query_tag
%type <nodes> args symbols

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
