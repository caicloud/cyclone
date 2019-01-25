import { Link } from 'react-router-dom';
import className from 'classnames/bind';
import PropTypes from 'prop-types';
import { Icon } from 'antd';

import styles from './list.module.less';

const clsStyles = className.bind(styles);

const KindFilter = ({ kinds, activeKind }) => (
  <ul className={styles['kind-list']}>
    {_.map(kinds, kind => {
      const itemCls = clsStyles('kind-item', {
        active: activeKind === kind.value,
      });
      return (
        <Link
          className={styles['kind-item-wrapper']}
          key={kind.value}
          to={{
            search: `?kind=${kind.value}`,
          }}
        >
          <li className={itemCls}>
            {kind.alias}
            <Icon type="right" className={styles['kind-item-ico']} />
          </li>
        </Link>
      );
    })}
  </ul>
);

KindFilter.propTypes = {
  kinds: PropTypes.arrayOf(
    PropTypes.shape({
      alias: PropTypes.string,
      value: PropTypes.string,
    })
  ).isRequired,
  activeKind: PropTypes.string,
};

export default KindFilter;
