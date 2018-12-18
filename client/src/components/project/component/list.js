import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('pipeline')
@observer
class List extends React.Component {
  render() {
    return <div>项目</div>;
  }
}

export default List;
