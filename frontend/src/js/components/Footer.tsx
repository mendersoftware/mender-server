// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import { makeStyles } from 'tss-react/mui';

const northernTechBlue = '#28aee4';

const useStyles = makeStyles()(theme => ({
  footer: {
    background: northernTechBlue,
    display: 'grid',
    minHeight: theme.mixins.toolbar.minHeight,
    gridTemplateColumns: '1fr max-content max-content',
    columnGap: theme.spacing(4),
    paddingLeft: '5vw',
    paddingRight: '5vw',
    alignItems: 'center',
    '>a': {
      color: '#fff'
    },
    '>a:hover': {
      color: '#092128'
    }
  }
}));

const companySite = 'https://northern.tech';

const getTargets = () => [
  { key: 'company', target: companySite, title: `© ${new Date().getFullYear()} Northern.tech` },
  { key: 'tos', target: `${companySite}/legal/hosted-mender-agreement-northern-tech-as.pdf`, title: 'Terms of service' },
  { key: 'privacyPolicy', target: `${companySite}/legal/privacy-policy`, title: 'Privacy policy' }
];

export const Footer = () => {
  const { classes } = useStyles();
  const targets = getTargets();
  return (
    <div className={classes.footer}>
      {targets.map(({ key, target, title }) => (
        /* eslint-disable-next-line react/jsx-no-target-blank */
        <a className="clickable" href={target} key={key} target="_blank" rel="noopener">
          {title}
        </a>
      ))}
    </div>
  );
};

export default Footer;
