# Lists and List Functions

Sprig provides a simple `list` type that can contain arbitrary sequential lists
of data. This is similar to arrays or slices, but lists are designed to be used
as immutable data types.

Create a list of integers:

```
$myList := list 1 2 3 4 5
```

The above creates a list of `[1 2 3 4 5]`.

## first

To get the head item on a list, use `first`.

`first $myList` returns `1`

## rest

To get the tail of the list (everything but the first item), use `rest`.

`rest $myList` returns `[2 3 4 5]`

## last

To get the last item on a list, use `last`:

`last $myList` returns `5`. This is roughly analogous to reversing a list and
then calling `first`.

## initial

This compliments `last` by returning all _but_ the last element.
`initial $myList` returns `[1 2 3 4]`.

## append

Append a new item to an existing list, creating a new list.

```
$new = append $myList 6
```

The above would set `$new` to `[1 2 3 4 5 6]`. `$myList` would remain unaltered.

## prepend

Push an alement onto the front of a list, creating a new list.

```
prepend $myList 0
```

The above would produce `[0 1 2 3 4 5]`. `$myList` would remain unaltered.

## reverse

Produce a new list with the reversed elements of the given list.

```
reverse $myList
```

The above would generate the list `[5 4 3 2 1]`.

## uniq

Generate a list with all of the duplicates removed.

```
list 1 1 1 2 | uniq
```

The above would produce `[1 2]`

## without

The `without` function filters items out of a list.

```
without $myList 3
```

The above would produce `[1 2 4 5]`

Without can take more than one filter:

```
without $myList 1 3 5
```

That would produce `[2 4]`

##  has

Test to see if a list has a particular element.

```
has $myList 4
```

The above would return `true`, while `has $myList "hello"` would return false.

## A Note on List Internals

A list is implemented in Go as a `[]interface{}`. For Go developers embedding
Sprig, you may pass `[]interface{}` items into your template context and be
able to use all of the `list` functions on those items.
