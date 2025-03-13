import { useEffect, useState } from 'react';
import { Item, fetchItems } from '~/api';

// const PLACEHOLDER_IMAGE = import.meta.env.VITE_FRONTEND_URL + '/logo192.png';

interface Prop {
  reload: boolean;
  onLoadCompleted: () => void;
}

export const ItemList = ({ reload, onLoadCompleted }: Prop) => {
  const [items, setItems] = useState<Item[]>([]);
  useEffect(() => {
    const fetchData = () => {
      fetchItems()
        .then((data) => {
          console.debug('GET success:', data);
          setItems(data.items);
          onLoadCompleted();
        })
        .catch((error) => {
          console.error('GET error:', error);
        });
    };

    if (reload) {
      fetchData();
    }
  }, [reload, onLoadCompleted]);

  return (
    <div className="ItemListContainer">
      {items?.map((item) => {
        return (
          <div key={item.id} className="ItemList">
            {/* TODO: Task 2: Show item images */}
            <img src={`http://localhost:9000/images/${item.image_name}.jpg`} />
            <p>
              <span className='itemName'>{item.name}</span>
              <br />
              <span>- {item.category}</span>
            </p>
          </div>
        );
      })}
    </div>
  );
};
