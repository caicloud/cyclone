import React from 'react';
import PropTypes from 'prop-types';
import { Menu, Dropdown, Icon } from 'antd';

class EllipsisMenu extends React.Component {
  static propTypes = {
    placement: PropTypes.oneOf([
      'bottomLeft',
      'bottomCenter',
      'bottomRight',
      'topLeft',
      'topCenter',
      'topRight',
    ]),
    menuText: PropTypes.oneOf([PropTypes.string, PropTypes.array]),
    menuFunc: PropTypes.oneOf([PropTypes.string, PropTypes.array]),
    disabled: PropTypes.oneOf([PropTypes.string, PropTypes.array]),
  };
  static defaultProps = {
    placement: 'bottomLeft',
    disabled: false,
    menuText: intl.get('operation.delete'),
  };
  render() {
    const { placement, menuText, menuFunc, disabled } = this.props;
    const menu = (
      <Menu>
        {_.isArray(menuText) ? (
          menuText.map((m, i) => (
            <Menu.Item key={m} onClick={menuFunc[i]} disabled={disabled[i]}>
              {m}
            </Menu.Item>
          ))
        ) : (
          <Menu.Item onClick={menuFunc} disabled={disabled}>
            {menuText}
          </Menu.Item>
        )}
      </Menu>
    );
    return (
      <Dropdown overlay={menu} placement={placement} trigger={['click']}>
        <Icon type="ellipsis" style={{ transform: 'rotate(90deg)' }} />
      </Dropdown>
    );
  }
}

export default EllipsisMenu;
