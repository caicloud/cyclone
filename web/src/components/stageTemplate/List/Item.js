import defaultCover from '@/images/stage/template_default.png';
import { Card, Tooltip, Modal, Icon } from 'antd';
import PropTypes from 'prop-types';
import { inject } from 'mobx-react';

import styles from './list.module.less';
const Fragment = React.Fragment;

const { Meta } = Card;

@inject('stageTemplate')
class Item extends React.Component {
  static propTypes = {
    template: PropTypes.shape({
      metadata: PropTypes.shape({
        name: PropTypes.string,
        annotations: PropTypes.object,
      }),
    }),
    history: PropTypes.object,
    stageTemplate: PropTypes.object,
  };

  state = {
    visible: false,
  };

  handleSubmit = name => {
    const { stageTemplate } = this.props;
    stageTemplate.deleteStageTemplate(name, () => {
      this.props.history.replace('/stageTemplate');
      stageTemplate.getTemplateList();
    });
  };

  render() {
    const { template, history } = this.props;
    const { visible } = this.state;
    const name = _.get(template, 'metadata.name');
    return (
      <Fragment>
        <Card
          onClick={() => {
            history.push(`/stageTemplate/${name}`);
          }}
          hoverable
          className={styles['template-item']}
          style={{ width: 208, margin: '0 16px 16px 0' }}
          cover={<img height="104px" alt="example" src={defaultCover} />}
          actions={[
            <Icon
              key="edit"
              type="edit"
              onClick={e => {
                e.stopPropagation();
                history.push(`/stageTemplate/add/${name}`);
              }}
            />,
            <Icon
              key="delete"
              type="delete"
              onClick={e => {
                e.stopPropagation();
                this.setState({ visible: true });
              }}
            />,
          ]}
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
        <Modal
          visible={visible}
          onCancel={() => this.setState({ visible: false })}
          onOk={() => {
            this.handleSubmit(name);
          }}
        >
          <p>{intl.get('template.deletetips')}</p>
        </Modal>
      </Fragment>
    );
  }
}

export default Item;
