import defaultCover from '@/images/stage/template_default.png';
import { Card, Tooltip } from 'antd';
import PropTypes from 'prop-types';

import styles from './list.module.less';

const { Meta } = Card;

const Item = ({ template }) => {
  return (
    <Card
      hoverable
      className={styles['template-item']}
      style={{ width: 208, height: 208, margin: '0 16px 16px 0' }}
      cover={<img height="104px" alt="example" src={defaultCover} />}
    >
      <Meta
        title={
          <Tooltip title={_.get(template, 'metadata.name')}>
            {_.get(template, 'metadata.name')}
          </Tooltip>
        }
        description={_.get(template, [
          'metadata',
          'annotations',
          'cyclone.io/description',
        ])}
      />
    </Card>
  );
};

Item.propTypes = {
  template: PropTypes.shape({
    metadata: PropTypes.shape({
      name: PropTypes.string,
      annotations: PropTypes.object,
    }),
  }),
};

export default Item;
