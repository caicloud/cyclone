import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('workflow')
@observer
class Template extends React.Component {
  render() {
    return <div>{intl.get('sideNav.template')}</div>;
  }
}

export default Template;
