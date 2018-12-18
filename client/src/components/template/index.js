import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('pipeline')
@observer
class Template extends React.Component {
  render() {
    return <div>模版</div>;
  }
}

export default Template;
