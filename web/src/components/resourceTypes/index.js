import React from 'react';
import { Route, Switch } from 'react-router-dom';
import List from './component/List';
import CreateOrUpdate from './component/create';
import Detail from './component/detail';
import PropTypes from 'prop-types';

const Resource = ({ match }) => {
  return (
    <Switch>
      <Route path="/resource" exact component={List} />
      <Route path={`${match.path}/add`} component={CreateOrUpdate} />
      <Route
        path={`${match.path}/:resourceTypeName/update`}
        component={CreateOrUpdate}
      />
      <Route path={`${match.path}/:resourceTypeName`} component={Detail} />
    </Switch>
  );
};

Resource.propTypes = {
  match: PropTypes.object,
};

export default Resource;
