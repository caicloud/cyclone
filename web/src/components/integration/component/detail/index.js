import { Spin, Button, Icon, Modal } from 'antd';
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
                _.get(detail, 'spec.cluster.isControlCluster') ? (
                  <Icon type="check" style={{ color: '#1890ff' }} />
                ) : (
                  <Icon type="minus" />
                )
              }
            />
            <DetailHeadItem
              name={intl.get('integration.form.cluster.isWorkerCluster')}
              value={
                _.get(detail, 'spec.cluster.isWorkerCluster') ? (
                  <Icon type="check" style={{ color: '#1890ff' }} />
                ) : (
                  <Icon type="minus" />
                )
              }
            />
          </div>
        );
      default:
        return <div />;
    }
  };

  openCluster = () => {
    const { integrationDetail, openCluster } = this.props.integration;
    Modal.confirm({
      title: intl.get('cluster.openTips'),
      onOk() {
        openCluster(integrationDetail.metadata.name);
      },
    });
  };

  closeCluster = () => {
    const { integrationDetail, closeCluster } = this.props.integration;
    Modal.confirm({
      title: intl.get('cluster.closeTips'),
      onOk() {
        closeCluster(integrationDetail.metadata.name);
      },
    });
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
    const processing = integration.processing;
    if (loading || processing) {
      return <Spin />;
    }
    const detail = integration.integrationDetail;
    const opened =
      _.get(
        detail,
        'metadata.labels["integration.cyclone.dev/schedulable-cluster"]'
      ) === 'true';

    return (
      <Detail
        actions={
          <DetailAction>
            {_.get(detail, 'spec.type') === 'Cluster' && (
              <Button
                type="primary"
                size="small"
                style={{ marginRight: 16 }}
                onClick={opened ? this.closeCluster : this.openCluster}
              >
                {opened ? intl.get('cluster.close') : intl.get('cluster.open')}
              </Button>
            )}
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
