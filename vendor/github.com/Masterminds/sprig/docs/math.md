# Math Functions

All math functions operate on `int64` values unless specified otherwise.

(In the future, these will be extended to handle floats as well)

## add

Sum numbers with `add`

## add1

To increment by 1, use `add1`

## sub

To subtract, use `sub`

## div

Perform integer division with `div`

## mod

Modulo with `mod`

## mul

Multiply with `mul`

## max

Return the largest of a series of integers:

This will return `3`:

```
max 1 2 3
```

## min

Return the smallest of a series of integers.

`min 1 2 3` will return `1`.

## floor

Returns the greatest float value less than or equal to input value

`floor 123.9999` will return `123.0`

## ceil

Returns the greatest float value greater than or equal to input value

`ceil 123.001` will return `124.0`

## round

Returns a float value with the remainder rounded to the given number to digits after the decimal point.

`round 123.555555` will return `123.556`