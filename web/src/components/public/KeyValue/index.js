import classNames from 'classnames/bind';
import { Tooltip, Icon } from 'antd';
import PropTypes from 'prop-types';

import style from './keyValue.module.less';

const clsStyle = classNames.bind(style);

// k-v (presentational component)
const KeyValue = ({ name, value, cls, tip, isEmpty }) => {
  const kvWrapperCls = clsStyle('item', 'kv-item', {
    [cls]: !!cls,
  });
  return (
    <div className={kvWrapperCls}>
      <div className={clsStyle('item-name', 'name')}>
        {name}
        {tip && (
          <Tooltip title={tip}>
            <Icon
              className={clsStyle('name-tip')}
              type="question-circle"
              theme="filled"
            />
          </Tooltip>
        )}
      </div>
      <div className={clsStyle('item-value', 'value')}>
        {/* display `--` while empty */}
        {isEmpty ? '--' : value}
      </div>
    </div>
  );
};

KeyValue.propTypes = {
  cls: PropTypes.string,
  name: PropTypes.string.isRequired,
  tip: PropTypes.string,
  value: PropTypes.node,
  isEmpty: PropTypes.bool,
};

export default KeyValue;
