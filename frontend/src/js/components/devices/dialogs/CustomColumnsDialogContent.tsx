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
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import ReactDOM from 'react-dom';

// material ui
import { Clear as ClearIcon, DragHandle as DragHandleIcon } from '@mui/icons-material';
import { DialogContent, FormControl, IconButton, ListItem } from '@mui/material';

import { attachClosestEdge, extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { getReorderDestinationIndex } from '@atlaskit/pragmatic-drag-and-drop-hitbox/util/get-reorder-destination-index';
import { DropIndicator } from '@atlaskit/pragmatic-drag-and-drop-react-drop-indicator/box';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { draggable, dropTargetForElements, monitorForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { pointerOutsideOfPreview } from '@atlaskit/pragmatic-drag-and-drop/element/pointer-outside-of-preview';
import { setCustomNativeDragPreview } from '@atlaskit/pragmatic-drag-and-drop/element/set-custom-native-drag-preview';
import { reorder } from '@atlaskit/pragmatic-drag-and-drop/reorder';
import { ATTRIBUTE_SCOPES } from '@northern.tech/store/constants';
import invariant from 'tiny-invariant';

import AttributeAutoComplete, { getOptionLabel } from '../widgets/AttributeAutocomplete';

type DraggableState = { type: 'idle' } | { type: 'preview'; container: HTMLElement } | { type: 'dragging' };

const idleState: DraggableState = { type: 'idle' };
const draggingState: DraggableState = { type: 'dragging' };

const DraggableListItem = ({ item, index, onRemove }) => {
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const title = useMemo(() => getOptionLabel(item), [item.key, item.scope, item.title]);
  const dragRef = useRef(null);
  const elementRef = useRef(null);
  const [draggableState, setDraggableState] = useState<DraggableState>(idleState);

  const [closestEdge, setClosestEdge] = useState(null);

  useEffect(() => {
    const dragHandle = dragRef.current;
    const element = elementRef.current;
    invariant(element);
    invariant(dragHandle);
    return combine(
      draggable({
        element: dragHandle,
        onDragStart: () => setDraggableState(draggingState),
        onDrop: () => setDraggableState(idleState),
        onGenerateDragPreview({ nativeSetDragImage }) {
          setCustomNativeDragPreview({
            nativeSetDragImage,
            getOffset: pointerOutsideOfPreview({
              x: '16px',
              y: '8px'
            }),
            render({ container }) {
              setDraggableState({ type: 'preview', container });
              return () => setDraggableState(draggingState);
            }
          });
        },
        getInitialData: () => ({ itemId: item.key })
      }),
      dropTargetForElements({
        element,
        getData: ({ input, element }) => {
          const data = { itemId: item.key };

          return attachClosestEdge(data, {
            input,
            element,
            allowedEdges: ['top', 'bottom']
          });
        },
        getIsSticky: () => false,
        onDragEnter: args => {
          if (args.source.data.itemId !== item.key) {
            setClosestEdge(extractClosestEdge(args.self.data));
          }
        },
        onDrag: args => {
          if (args.source.data.itemId !== item.key) {
            setClosestEdge(extractClosestEdge(args.self.data));
          }
        },
        onDragLeave: () => {
          setClosestEdge(null);
        },
        onDrop: () => {
          setClosestEdge(null);
        }
      })
    );
  }, [item.key]);
  const onClick = () => onRemove(item, index);
  return (
    <div ref={elementRef} className="relative">
      <ListItem className={`flexbox space-between margin-right-large ${draggableState.type === 'dragging' ? 'dragging' : ''}`}>
        <div>{title}</div>
        <div className="flexbox space-between" style={{ width: 80 }}>
          <div className="flexbox centered cursor-grab" ref={dragRef}>
            <DragHandleIcon />
          </div>
          <IconButton onClick={onClick} size="small">
            <ClearIcon color="disabled" />
          </IconButton>
        </div>
      </ListItem>
      {closestEdge && <DropIndicator edge={closestEdge} />}
      {draggableState.type === 'preview' && ReactDOM.createPortal(<div>{title}</div>, draggableState.container)}
    </div>
  );
};

const filterAttributes = (list, attribute) => list.filter(item => !(item.key === attribute.key && item.scope === attribute.scope));

const Content = ({ attributes, columnHeaders, idAttribute, selectedAttributes, setSelectedAttributes }) => {
  const [attributeOptions, setAttributeOptions] = useState([]);

  useEffect(() => {
    const { attributeOptions, selectedAttributes } = columnHeaders.reduce(
      (accu, { attribute, title }, index) => {
        // we skip the first/ id column + exclude the status column from customization
        if (index && attribute.name && !(attribute.name === 'status' && attribute.scope === ATTRIBUTE_SCOPES.identity)) {
          const currentAttribute = { ...attribute, key: attribute.name, id: `${attribute.scope}-${attribute.name}`, title };
          accu.selectedAttributes.push(currentAttribute);
          accu.attributeOptions = filterAttributes(accu.attributeOptions, currentAttribute);
        }
        return accu;
      },
      {
        attributeOptions: [...attributes.filter(item => !([idAttribute.attribute, 'status'].includes(item.key) && item.scope === ATTRIBUTE_SCOPES.identity))],
        selectedAttributes: []
      }
    );
    setSelectedAttributes(selectedAttributes);
    setAttributeOptions(attributeOptions);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(attributes), JSON.stringify(columnHeaders), idAttribute.attribute, setSelectedAttributes]);

  const reorderItem = useCallback(
    ({ startIndex, finishIndex }) => {
      const updatedItems = reorder({
        list: selectedAttributes,
        startIndex,
        finishIndex
      });
      setSelectedAttributes(updatedItems);
    },
    [selectedAttributes, setSelectedAttributes]
  );
  useEffect(() => {
    return monitorForElements({
      onDrop: ({ source, location }) => {
        if (!location.current.dropTargets.length) {
          return;
        }
        const draggedItemId = source.data.itemId;
        const [destinationItem] = location.current.dropTargets;
        const draggedItemIndex = selectedAttributes.findIndex(item => item.key === draggedItemId);
        const indexOfTarget = selectedAttributes.findIndex(item => item.key === destinationItem.data.itemId);
        const closestEdgeOfTarget = extractClosestEdge(destinationItem.data);
        const destinationIndex = getReorderDestinationIndex({
          startIndex: draggedItemIndex,
          indexOfTarget: indexOfTarget,
          closestEdgeOfTarget: closestEdgeOfTarget,
          axis: 'vertical'
        });
        reorderItem({
          startIndex: draggedItemIndex,
          finishIndex: destinationIndex
        });
      }
    });
  }, [selectedAttributes, reorderItem]);

  const onRemove = (attribute, index) => {
    let selection = [];
    let removed = attribute;
    if (index !== undefined) {
      selection = [...selectedAttributes];
      const [removedAttribute] = selection.splice(index, 1);
      removed = removedAttribute;
    } else {
      selection = filterAttributes(selectedAttributes, attribute);
    }
    setSelectedAttributes(selection);
    setAttributeOptions([...attributeOptions, removed]);
  };

  const onSelect = useCallback(
    attribute => {
      if (attribute.key) {
        const existingAttribute = attributeOptions.find(item => item.key === attribute.key && item.scope === attribute.scope) || attribute;
        setSelectedAttributes(current => [
          ...current,
          { ...existingAttribute, title: existingAttribute.value ?? existingAttribute.key, id: `${attribute.scope}-${attribute.key}` }
        ]);
        setAttributeOptions(filterAttributes(attributeOptions, attribute));
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(attributeOptions), setSelectedAttributes]
  );

  return (
    <DialogContent>
      <p>You can select columns of inventory data to display in the device list table. Drag to change the order.</p>
      <div>
        {selectedAttributes.map((item, index) => (
          <DraggableListItem item={item} index={index} key={item.key} onRemove={onRemove} />
        ))}
      </div>
      <FormControl>
        <AttributeAutoComplete attributes={attributeOptions} label="Add a column" onRemove={onRemove} onSelect={onSelect} />
      </FormControl>
    </DialogContent>
  );
};

export default Content;
