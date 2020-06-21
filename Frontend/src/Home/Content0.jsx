import React from 'react';
import QueueAnim from 'rc-queue-anim';
import { Row, Col,Button,Card } from 'antd';
import OverPack from 'rc-scroll-anim/lib/ScrollOverPack';
import { getChildrenToRender } from './utils';

class Content extends React.PureComponent {
  render() {
    const { dataSource, isMobile,callBack, ...props } = this.props;
    const {
      wrapper,
      titleWrapper,
      page,
      OverPack: overPackData,
      childWrapper,
    } = dataSource;
    return (
      <div {...props} {...wrapper}>
        <div {...page}>
          <div {...titleWrapper}>
            {titleWrapper.children.map(getChildrenToRender)}
          </div>
          <OverPack {...overPackData}>
            <QueueAnim
              type="bottom"
              key="block"
              leaveReverse
              component={Row}
              componentProps={childWrapper}
            >
              {childWrapper.children.map((block, i) => {
                const { children: item, ...blockProps } = block;
                return (
                  <Col onClick={()=>callBack(i.toString())} {...blockProps}>
                    <Card>
                    <div {...item}>
                      {item.children.map(getChildrenToRender)}
                    </div>
                    </Card>
                   
                  </Col>
                );
              })}
            </QueueAnim>
          </OverPack>
        </div>
      </div>
    );
  }
}

export default Content;
