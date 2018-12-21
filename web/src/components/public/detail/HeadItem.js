import React from 'react';
import PropTypes from 'prop-types';

const DetailHeadItem = props => {
  const { name, value } = props;
  return (
    <div className="head-item">
      <div className="name">{name}</div>
      <div className="value">{value}</div>
    </div>
  );
};

DetailHeadItem.propTypes = {
  className: PropTypes.string,
  onClick: PropTypes.func,
  value: PropTypes.any,
  name: PropTypes.any,
};

export default DetailHeadItem;
