import { useRef, useState } from 'react';
import { postItem } from '~/api';
import './Listing.css';


interface Prop {
  onListingCompleted: () => void;
}

type FormDataType = {
  name: string;
  category: string;
  image: File | null;
};

export const Listing = ({ onListingCompleted }: Prop) => {
  const initialState = {
    name: '',
    category: '',
    image: null,
  };
  const [values, setValues] = useState<FormDataType>(initialState);

  const uploadImageRef = useRef<HTMLInputElement>(null);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.value,
    });
  };
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.files![0] || null,
    });
  };
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    // Validate field before submit
    const REQUIRED_FILEDS = ['name'];
    const missingFields = Object.entries(values)
      .filter(([key, value]) => !value && REQUIRED_FILEDS.includes(key))
      .map(([key]) => key);

    if (missingFields.length) {
      alert(`Missing fields: ${missingFields.join(', ')}`);
      return;
    }

    // Submit the form
    postItem({
      name: values.name,
      category: values.category,
      image: values.image,
    })
      .then(() => {
        alert('Item listed successfully');
      })
      .catch((error) => {
        console.error('POST error:', error);
        alert('Failed to list this item');
      })
      .finally(() => {
        onListingCompleted();
        setValues(initialState);
        if (uploadImageRef.current) {
          uploadImageRef.current.value = '';
        }
      });
  };
  return (
    <div className="Listing">
      <form onSubmit={onSubmit}>
        <div>
          <input
            type="text"
            name="name"
            id="name"
            placeholder="name"
            onChange={onValueChange}
            required
            value={values.name}
          />
          <input
            type="text"
            name="category"
            id="category"
            placeholder="category"
            onChange={onValueChange}
            value={values.category}
          />
          <input
            type="file"
            name="image"
            id="image"
            onChange={onFileChange}
            ref={uploadImageRef}
            required
          />
          <button type="submit">List this item</button>
        </div>
      </form>
    </div>
  );
};
