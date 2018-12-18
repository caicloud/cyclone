import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('pipeline')
@observer
class Overview extends React.Component {
  render() {
    return <div>总览</div>;
  }
}

export default Overview;
