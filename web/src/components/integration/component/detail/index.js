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

  detailContent = detail => {
    const type = _.get(detail, 'spec.type');
    switch (type) {
      case 'SCM':
        return (
          <div style={{ paddingBottom: 16 }}>
            <DetailHeadItem
              name={intl.get('integration.form.scm.type')}
              value={_.get(detail, 'spec.scm.type')}
            />
            <DetailHeadItem
              name={intl.get('integration.form.scm.serverAddress')}
              value={_.get(detail, 'spec.scm.server')}
            />
            <DetailHeadItem
              name={intl.get('integration.form.scm.authType')}
              value={
                _.get(detail, 'spec.scm.authType') === 'Password'
                  ? intl.get('integration.form.scm.usernamepwd')
                  : 'Token'
              }
            />
            {_.get(detail, 'spec.scm.type') && (
              <DetailHeadItem
                name={intl.get('integration.form.username')}
                value={_.get(detail, 'spec.scm.user')}
              />
            )}
          </div>
        );
      case 'DockerRegistry':
        return (
          <div style={{ paddingBottom: 16 }}>
            <DetailHeadItem
              name={intl.get('integration.form.dockerRegistry.registryAddress')}
              value={_.get(detail, 'spec.dockerRegistry.server')}
            />
            <DetailHeadItem
              name={intl.get('integration.form.username')}
              value={_.get(detail, 'spec.dockerRegistry.user')}
            />
          </div>
        );
      case 'Cluster':
        return (
          <div style={{ paddingBottom: 16 }}>
            <DetailHeadItem
              name={intl.get('integration.form.cluster.isControlCluster')}
              value={
                _.get(detail, 'spec.cluster.isControlCluster') ? 'YES' : 'NO'
              }
            />
            <DetailHeadItem
              name={intl.get('integration.form.cluster.isWorkerCluster')}
              value={
                _.get(detail, 'spec.cluster.isWorkerCluster') ? 'YES' : 'NO'
              }
            />
          </div>
        );
      default:
        return <div />;
    }
  };

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
          <DetailHeadItem
            name={intl.get('integration.type')}
            value={_.get(detail, 'spec.type')}
          />
          <DetailHeadItem
            name={intl.get('integration.desc')}
            value={_.get(detail, 'metadata.description') || ' -- '}
          />
          {this.detailContent(detail)}
        </DetailHead>
      </Detail>
    );
  }
}

export default IntegrationDetail;
