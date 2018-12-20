import React from 'react';
import PropTypes from 'prop-types';

const DetailContent = props => {
  const { children } = props;
  return <div className="detail-content">{children}</div>;
};
DetailContent.propTypes = {
  children: PropTypes.node,
};
export default DetailContent;
