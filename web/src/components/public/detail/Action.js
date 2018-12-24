import React from 'react';
import PropTypes from 'prop-types';

const Action = props => {
  const { children } = props;
  return <div className="action">{children}</div>;
};
Action.propTypes = {
  children: PropTypes.node,
};
export default Action;
