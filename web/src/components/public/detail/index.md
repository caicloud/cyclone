# Detail

```js
import Detail from '@/public/detail';
import { Button } from 'antd';

const { DetailHead, DetailHeadItem, DetailContent, DetailAction } = Detail;

<Detail
  actions={
    <DetailAction>
      <Button>{intl.get('operation.update')}</Button>
    </DetailAction>
  }
>
  <DetailHead headName="header name">
    <DetailHeadItem name="key" value="2018-09-08" />
  </DetailHead>
  <DetailContent>content</DetailContent>
</Detail>;
```
