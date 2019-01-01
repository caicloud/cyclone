import PropTypes from 'prop-types';
import React from 'react';
import { Row } from 'antd';

import Item from './Item';

const List = ({ list }) => {
  return _.map(list, (chunk, index) => (
    <Row gutter={16} key={index} style={{ marginBottom: 16 }}>
      {_.map(chunk, template => (
        <Item template={template} key={_.get(template, 'metadata.name')} />
      ))}
    </Row>
  ));
};

List.propTypes = {
  list: PropTypes.array,
};

export default List;
