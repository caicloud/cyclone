import Detail from '@/components/public/detail';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { Tabs, Spin } from 'antd';

import Inputs from './Inputs';
import Outputs from './Outputs';
import Configuration from './Configuration';

const { DetailHead, DetailHeadItem, DetailContent, DetailAction } = Detail;
const TabPane = Tabs.TabPane;

@inject('stageTemplate')
@observer
class TemplateDetail extends React.Component {
  componentDidMount() {
    const {
      match: {
        params: { templateName },
      },
    } = this.props;
    templateName && this.props.stageTemplate.getTemplate(templateName);
  }

  render() {
    const {
      match: {
        params: { templateName },
      },
      stageTemplate: { template, templateLoading },
    } = this.props;
    if (_.get(template, 'metadata.name') !== templateName || templateLoading) {
      return <Spin />;
    }
    return (
      <Detail actions={<DetailAction />}>
        <DetailHead headName={templateName}>
          <DetailHeadItem
            name={intl.get('creationTime')}
            value={_.get(template, 'metadata.creationTimestamp')}
          />
          <DetailHeadItem
            name={intl.get('description')}
            value={
              _.get(template, [
                'metadata',
                'annotations',
                'cyclone.io/description',
              ]) || '--'
            }
          />
        </DetailHead>
        <DetailContent>
          <Tabs defaultActiveKey="inputs" type="card">
            <TabPane tab={intl.get('stage.inputs')} key="inputs">
              <Inputs inputs={_.get(template, 'spec.pod.inputs')} />
            </TabPane>
            <TabPane tab={intl.get('stage.outputs')} key="outputs">
              <Outputs outputs={_.get(template, 'spec.pod.outputs')} />
            </TabPane>
            <TabPane tab={intl.get('configuration')} key="configuration">
              <Configuration configuration={_.get(template, 'spec.pod.spec')} />
            </TabPane>
          </Tabs>
        </DetailContent>
      </Detail>
    );
  }
}

TemplateDetail.propTypes = {
  match: PropTypes.shape({
    params: PropTypes.shape({
      templateName: PropTypes.string,
    }).isRequired,
  }),
  stageTemplate: PropTypes.shape({
    getTemplate: PropTypes.func,
    template: PropTypes.object,
    templateLoading: PropTypes.bool,
  }),
};

export default TemplateDetail;
