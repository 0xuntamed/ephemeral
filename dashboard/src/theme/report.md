# Colour palette report

- Source brand: `#B8321F`
- Source HSL: `7.45deg 71.16% 42.16%`
- Nearest derived shade: `600` (`#AF2F1D`)
- Text target: WCAG AA

The exact source colour is preserved as `--color-source-brand`; the shade scale is derived from its hue and saturation.

## Shade scale

| Shade | Hex       | HSL                  |
| ----: | --------- | -------------------- |
|    50 | `#FCF4F3` | `6.7deg 60.0% 97.1%` |
|   100 | `#F8E9E7` | `7.1deg 54.8% 93.9%` |
|   200 | `#F2CFCA` | `7.5deg 60.6% 87.1%` |
|   300 | `#E8A196` | `8.0deg 64.1% 74.9%` |
|   400 | `#E06D5D` | `7.3deg 67.9% 62.2%` |
|   500 | `#D23923` | `7.5deg 71.4% 48.0%` |
|   600 | `#AF2F1D` | `7.4deg 71.6% 40.0%` |
|   700 | `#902718` | `7.5deg 71.4% 32.9%` |
|   800 | `#762014` | `7.3deg 71.0% 27.1%` |
|   900 | `#57180F` | `7.5deg 70.6% 20.0%` |
|   950 | `#2C0C07` | `8.1deg 72.5% 10.0%` |

## Light mode tokens

| Token                    | Value     | Source      |
| ------------------------ | --------- | ----------- |
| `background`             | `#FFFFFF` | `#FFFFFF`   |
| `foreground`             | `#2C0C07` | `shade:950` |
| `card`                   | `#FFFFFF` | `#FFFFFF`   |
| `card-foreground`        | `#57180F` | `shade:900` |
| `popover`                | `#FFFFFF` | `#FFFFFF`   |
| `popover-foreground`     | `#2C0C07` | `shade:950` |
| `primary`                | `#AF2F1D` | `shade:600` |
| `primary-foreground`     | `#FFFFFF` | `#FFFFFF`   |
| `secondary`              | `#F8E9E7` | `shade:100` |
| `secondary-foreground`   | `#57180F` | `shade:900` |
| `muted`                  | `#FCF4F3` | `shade:50`  |
| `muted-foreground`       | `#AF2F1D` | `shade:600` |
| `accent`                 | `#F8E9E7` | `shade:100` |
| `accent-foreground`      | `#57180F` | `shade:900` |
| `destructive`            | `#B91C1C` | `#B91C1C`   |
| `destructive-foreground` | `#FFFFFF` | `#FFFFFF`   |
| `border`                 | `#E06D5D` | `shade:400` |
| `input`                  | `#E06D5D` | `shade:400` |
| `ring`                   | `#D23923` | `shade:500` |

## Dark mode tokens

| Token                    | Value     | Source      |
| ------------------------ | --------- | ----------- |
| `background`             | `#2C0C07` | `shade:950` |
| `foreground`             | `#FCF4F3` | `shade:50`  |
| `card`                   | `#57180F` | `shade:900` |
| `card-foreground`        | `#FCF4F3` | `shade:50`  |
| `popover`                | `#57180F` | `shade:900` |
| `popover-foreground`     | `#FCF4F3` | `shade:50`  |
| `primary`                | `#E06D5D` | `shade:400` |
| `primary-foreground`     | `#2C0C07` | `shade:950` |
| `secondary`              | `#762014` | `shade:800` |
| `secondary-foreground`   | `#FCF4F3` | `shade:50`  |
| `muted`                  | `#762014` | `shade:800` |
| `muted-foreground`       | `#E8A196` | `shade:300` |
| `accent`                 | `#762014` | `shade:800` |
| `accent-foreground`      | `#FCF4F3` | `shade:50`  |
| `destructive`            | `#EF4444` | `#EF4444`   |
| `destructive-foreground` | `#2C0C07` | `shade:950` |
| `border`                 | `#D23923` | `shade:500` |
| `input`                  | `#D23923` | `shade:500` |
| `ring`                   | `#D23923` | `shade:500` |

## Contrast checks

| Mode  | Check                    | Pair                                     |   Ratio | Minimum | Result   |
| ----- | ------------------------ | ---------------------------------------- | ------: | ------: | -------- |
| light | body text                | `foreground` / `background`              | 18.06:1 |   4.5:1 | **PASS** |
| light | card text                | `card-foreground` / `card`               | 13.61:1 |   4.5:1 | **PASS** |
| light | popover text             | `popover-foreground` / `popover`         | 18.06:1 |   4.5:1 | **PASS** |
| light | primary control text     | `primary-foreground` / `primary`         |  6.47:1 |   4.5:1 | **PASS** |
| light | primary as link          | `primary` / `background`                 |  6.47:1 |   4.5:1 | **PASS** |
| light | secondary control text   | `secondary-foreground` / `secondary`     | 11.54:1 |   4.5:1 | **PASS** |
| light | muted text               | `muted-foreground` / `muted`             |  5.97:1 |   4.5:1 | **PASS** |
| light | accent text              | `accent-foreground` / `accent`           | 11.54:1 |   4.5:1 | **PASS** |
| light | destructive control text | `destructive-foreground` / `destructive` |  6.47:1 |   4.5:1 | **PASS** |
| light | border on page           | `border` / `background`                  |  3.22:1 |   3.0:1 | **PASS** |
| light | input boundary on page   | `input` / `background`                   |  3.22:1 |   3.0:1 | **PASS** |
| light | focus ring on page       | `ring` / `background`                    |  4.83:1 |   3.0:1 | **PASS** |
| dark  | body text                | `foreground` / `background`              | 16.66:1 |   4.5:1 | **PASS** |
| dark  | card text                | `card-foreground` / `card`               | 12.56:1 |   4.5:1 | **PASS** |
| dark  | popover text             | `popover-foreground` / `popover`         | 12.56:1 |   4.5:1 | **PASS** |
| dark  | primary control text     | `primary-foreground` / `primary`         |   5.6:1 |   4.5:1 | **PASS** |
| dark  | primary as link          | `primary` / `background`                 |   5.6:1 |   4.5:1 | **PASS** |
| dark  | secondary control text   | `secondary-foreground` / `secondary`     |  9.75:1 |   4.5:1 | **PASS** |
| dark  | muted text               | `muted-foreground` / `muted`             |  5.02:1 |   4.5:1 | **PASS** |
| dark  | accent text              | `accent-foreground` / `accent`           |  9.75:1 |   4.5:1 | **PASS** |
| dark  | destructive control text | `destructive-foreground` / `destructive` |   4.8:1 |   4.5:1 | **PASS** |
| dark  | border on page           | `border` / `background`                  |  3.74:1 |   3.0:1 | **PASS** |
| dark  | input boundary on page   | `input` / `background`                   |  3.74:1 |   3.0:1 | **PASS** |
| dark  | focus ring on page       | `ring` / `background`                    |  3.74:1 |   3.0:1 | **PASS** |

Required checks: **PASS** (0 failures).
