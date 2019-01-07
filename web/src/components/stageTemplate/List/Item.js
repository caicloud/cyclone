import defaultCover from '@/images/stage/template_default.png';

import PropTypes from 'prop-types';
import { Col, Card } from 'antd';
import React from 'react';

const { Meta } = Card;

const Item = ({ template }) => {
  return (
    <Col
      span={6}
      style={{
        display: 'flex',
        justifyContent: 'center',
        marginBottom: '16px',
      }}
    >
      <Card
        hoverable
        style={{ width: 240, height: 240 }}
        cover={<img height="120px" alt="example" src={defaultCover} />}
      >
        <Meta
          title={_.get(template, 'metadata.name')}
          description={_.get(template, 'metadata.description')}
        />
      </Card>
    </Col>
  );
};

Item.propTypes = {
  template: PropTypes.shape({
    metadata: {
      name: PropTypes.string,
      description: PropTypes.string,
    },
  }),
};

export default Item;
