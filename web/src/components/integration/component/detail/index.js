import { Spin } from 'antd';
import { inject, observer } from 'mobx-react';
import Detail from '@/components/public/detail';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from '../MenuAction';

const { DetailHead, DetailHeadItem, DetailAction } = Detail;

@inject('integration')
@observer
class IntegrationDetail extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    integration: PropTypes.object,
    history: PropTypes.object,
  };
  constructor(props) {
    super(props);
    const {
      match: {
        params: { integrationName },
      },
    } = this.props;
    this.props.integration.getIntegration(integrationName);
  }
  render() {
    const {
      integration,
      match: {
        params: { integrationName },
      },
      history,
    } = this.props;
    const loading = integration.detailLoading;
    if (loading) {
      return <Spin />;
    }
    const detail = integration.integrationDetail;

    return (
      <Detail
        actions={
          <DetailAction>
            <MenuAction name={integrationName} history={history} detail />
          </DetailAction>
        }
      >
        <DetailHead
          headName={_.get(detail, 'metadata.annotations["cyclone.dev/alias"]')}
        >
          <DetailHeadItem
            name={intl.get('creationTime')}
            value={FormatTime(_.get(detail, 'metadata.creationTimestamp'))}
          />
        </DetailHead>
      </Detail>
    );
  }
}

export default IntegrationDetail;
