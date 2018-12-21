# EllipsisMenu

### single delete action

```js
import EllipsisMenu from '@/public/ellipsisMenu';

<EllipsisMenu menuFunc={() => {}} disabled>

```

### Multiple drop-down actions

```js
import EllipsisMenu from '@/public/ellipsisMenu';

<EllipsisMenu menuFunc={[() => {}, () => {}]} menuText={[intl.get("operation.modify"),intl.get("operation.delete")]} disabled={[false, true]}>

```
