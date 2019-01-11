import PropTypes from 'prop-types';
import React from 'react';

import Item from './Item';

const List = ({ list }) => {
  return (
    <div
      style={{
        marginBottom: 16,
        display: 'flex',
        flexFlow: 'row wrap',
        alignContent: 'space-between',
        justifyContent: 'flex-start',
      }}
    >
      {_.map(list, template => (
        <Item template={template} key={_.get(template, 'metadata.name')} />
      ))}
    </div>
  );
};

List.propTypes = {
  list: PropTypes.array,
};

export default List;
