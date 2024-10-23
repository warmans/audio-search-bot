### Queries

```
| Prefix | Field          | Example                 | Description                               |
|--------|----------------|-------------------------|-------------------------------------------|
| ~      | publication    | `~sunny`                | Filters subtitles by publication          |
| #      | series/episode | `#S1E04`, `#S1`, `#E04` | Filter by a series and/or episode number. |
| +      | timestamp      | `+1m`, `+10m30s`        | Filter by timestamp greater than.         |
| "      | content        | `"day man"`             | Phrase match                              |
```

__Examples__

* `day man` - search for any dialog containing `day` or `man` in any order/location.
* `"day man"` - search for any dialog containing the phrase `day man` in that order (case insensitive).
* `~sunny day` - search for any dialog from the `sunny` publication containing `day`.
* `~sunny +1m30s #S3E09 man "day"` - search for dialog from the `sunny` publication, season 3 episode 9 occurring after `1m30s` and containing the word `man` and `day`.

### Paging

You can page results with the `>` operator in a query e.g. `>10`.

* `man >20` - search for dialog containing `man`, but skip the first 20 results.
* `~sunny +1m30s #S3E09 man "day" >100` - complex query with the fist 100 results skipped. 