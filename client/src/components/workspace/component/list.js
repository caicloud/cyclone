import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('pipeline')
@observer
class List extends React.Component {
  render() {
    return <div>Workspace</div>;
  }
}

export default List;
