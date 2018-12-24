import React from 'react';
import PropTypes from 'prop-types';

const DetailHead = props => {
  const { children, headName } = props;
  return (
    <div className="detail-head">
      {headName && <div className="head-name">{headName}</div>}
      {children}
    </div>
  );
};
DetailHead.propTypes = {
  children: PropTypes.node,
  headName: PropTypes.string,
};
export default DetailHead;
