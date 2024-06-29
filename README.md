# GQLGen
Gqlgen generates strings that conform to the XML encoded BNF grammar contained in the [GQL Standard](https://www.iso.org/standard/76120.html).

# Building

Gqlgen is written in Go. Follow the [Go installation instructions](https://go.dev/doc/install) to install Go.

Run the following command from the Gqlgen source directory.

```
go build
```

This will build Gqlgen and install the executable in the Gqlgen source directory. The program will be named "gqlgen" on Unix varients, including OS X or will be name "gqlgen.exe" on Windows

# Running Gqlgen

## Arguments

| Argument Name | Type    | Default     | Description                                             |
| ------------- |---------|-------------|---------------------------------------------------------|
| bnf | string  | ./bnf.xml   | Filename of the XML file containing the GQL BNF Grammar |
| start | string | GQL-program | Start rule name                                         |
| cnt | integer | 1           | Number of strings to generate                           |
| v | none    | false       | If specified run in verbose mode                        |

Gqlgen generates strings that conform to the whole language or to subsets of the language. The `start` parameter specifies the rule name to use as the starting point for generating strings. For example, to generate <value expressions>s, set start to "value expression", as is shown in the following example.

```
./gqlgen -start "value expression"
```

To generate more than one string at a time, use the `cnt` argument to specify the number of strings, as show in the following example.

```
./gqlgen -start "value expression" -cnt 5
```

Gqlgen makes random choices to decide which altertives to use, whether to include an optional rule, and how many times to repeat repeated rules. To avoid endless recursion a limit of 6 is used. When that limit is exceeded, Gqlgen retries the string creation.

Currently, Gqlgen is rather naive in terms of picking alternatives. This leads to the case that Gqlgen rarely chooses some alternatives. One critical example is that Gqlgen usually chooses Session commands over procedure definitions. The most effective way to use Gqlgen is to specify a start rule. For example, the following command generates 5 <value expression>s. The results are shown.

```
./gqlgen -start "value expression" -cnt 5

 TABLE "LòD`Z¶Xxá`4sXv`ŹzPP0LUj4UXòH`zlOnòVBuŹ_8OŨáváD"
Ues IS  TYPED  PATH   NOT  NULL  AND  NOT ( NOT  ALL_DIFFERENT (w,DCkO) IS  TRUE ) IS  TRUE 
uTrU0
 SAME (qka,zC,jj9Up,W) IS  NOT  FALSE  OR  NOT EXZ IS  SOURCE  OF LevjU IS  NOT  FALSE  AND  NOT w IS  DESTINATION  OF pX IS  NOT  UNKNOWN 
q
```

Each string is separated by a newline.