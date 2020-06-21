/* eslint no-undef: 0 */
/* eslint arrow-parens: 0 */
import React from 'react';
import { enquireScreen } from 'enquire-js';
import { animateScroll } from "react-scroll";

import Nav0 from './Nav0';
import Banner0 from './Banner0';
import Content0 from './Content0';
import Content1 from './Content1';
import Content3 from './Content3';
import Footer0 from './Footer0';

import {
  Nav00DataSource,
  Banner00DataSource,
  Content00DataSource,
  Content10DataSource,
  Content101DataSource,
  Content102DataSource,
  Content30DataSource,
  Footer01DataSource,
} from './data.source';
import './less/antMotionStyle.less';

let isMobile;
enquireScreen((b) => {
  isMobile = b;
});

const { location } = window;

export default class Home extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isMobile,
      show: !location.port, // 如果不是 dva 2.0 请删除
      isIGActive:false,
      isTitokActive:false,
      isFbActive:false,
    };
    this.myRef = React.createRef()  
  }
  componentDidMount() {
    // 适配手机屏幕;
    enquireScreen((b) => {
      this.setState({ isMobile: !!b });
    });
    // dva 2.0 样式在组件渲染之后动态加载，导致滚动组件不生效；线上不影响；
    /* 如果不是 dva 2.0 请删除 start */
    if (location.port) {
      // 样式 build 时间在 200-300ms 之间;
      setTimeout(() => {
        this.setState({
          show: true,
        });
      }, 500);
    }
    /* 如果不是 dva 2.0 请删除 end */
  }
  scrollToBottom() {
    animateScroll.scrollToBottom({
      containerId: "end",
      offset:10000
    });
}
  changeActive = (id)=>{
    if (id === "0"){
      this.setState({
        isIGActive:false,
        isTitokActive:true,
        isFbActive:false,
      })
    }
    if (id === "1"){
      this.setState({
        isIGActive:true,
        isTitokActive:false,
        isFbActive:false,
      })
    }
    if (id === "2"){
      this.setState({
        isIGActive:false,
        isTitokActive:false,
        isFbActive:true,
      })
    }
    this.myRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
  render() {
    const children = [
      <Banner0
        id="Banner0_0"
        key="Banner0_0"
        dataSource={Banner00DataSource}
        isMobile={this.state.isMobile}
      />,
      <Content0
      id="Cobtend0"
      key="Conttend 1"
      dataSource={Content00DataSource}
      isMobile={this.state.isMobile}
      callBack={(id)=>this.changeActive(id)}
      />,
    ];
    return (
      <div
        className="templates-wrapper"
        ref={(d) => {
          this.dom = d;
        }}
      >
        {/* 如果不是 dva 2.0 替换成 {children} start */}
        {this.state.show && children}
        <Content1
      id="tiktok"
      key="tiktok"
      dataSource={Content10DataSource}
      isMobile={this.state.isMobile}
      hide={!this.state.isTitokActive}
      kind={0}
      />,
      <Content1
      id="instagram"
      key="instagram"
      dataSource={Content101DataSource}
      isMobile={this.state.isMobile}
      hide={!this.state.isIGActive}
      kind={1}
      />,
      <Content1
      id="facebook"
      key="facebook"
      dataSource={Content102DataSource}
      isMobile={this.state.isMobile}
      hide={!this.state.isFbActive}
      kind={2}
            />,
        {/* 如果不是 dva 2.0 替换成 {children} end */}
        <div ref={this.myRef} id="end"></div>
      </div>
    );
  }
}
