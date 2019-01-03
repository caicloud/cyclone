import React from 'react';
import PropTypes from 'prop-types';

const Action = props => {
  return <div className="action">{props.children}</div>;
};
Action.propTypes = {
  children: PropTypes.node,
};
export default Action;
