import React from 'react';
import { inject, observer } from 'mobx-react';
@inject('workflow')
@observer
class List extends React.Component {
  render() {
    return <div>{intl.get('sideNav.project')}</div>;
  }
}

export default List;
