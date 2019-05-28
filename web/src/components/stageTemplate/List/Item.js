import defaultCover from '@/images/stage/template_default.png';
import { Card, Tooltip, Popconfirm, Icon } from 'antd';
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
    key: PropTypes.string,
    stageTemplate: PropTypes.object,
  };

  handleDelete = name => {
    const { stageTemplate } = this.props;
    stageTemplate.deleteStageTemplate(name, () => {
      this.props.history.replace('/stageTemplate');
      stageTemplate.getTemplateList();
    });
  };

  render() {
    const { template, history, key } = this.props;
    const actionList = [
      <Icon
        key="edit"
        type="edit"
        onClick={e => {
          e.stopPropagation();
          history.push(`/stageTemplate/${name}/update`);
        }}
      />,
      <Popconfirm
        key={key}
        title={intl.get('template.deletetips')}
        onConfirm={e => {
          e.stopPropagation();
          this.handleDelete(name);
        }}
        onCancel={e => {
          e.stopPropagation();
        }}
        okText={intl.get('confirm')}
        cancelText={intl.get('cancel')}
      >
        <Icon
          key="delete"
          type="delete"
          onClick={e => {
            e.stopPropagation();
          }}
        />
      </Popconfirm>,
    ];
    const name = _.get(template, 'metadata.name');
    return (
      <Fragment>
        <Card
          onClick={() => {
            history.push(`/stageTemplate/${name}`);
          }}
          hoverable
          className={styles['template-item']}
          cover={<img alt="template icon" src={defaultCover} />}
          actions={actionList}
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
              'cyclone.dev/description',
            ])}
          />
        </Card>
      </Fragment>
    );
  }
}

export default Item;
