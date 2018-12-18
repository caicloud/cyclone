import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('pipeline')
@observer
class Resource extends React.Component {
  render() {
    return <div>资源</div>;
  }
}

export default Resource;
