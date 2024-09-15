import React, { useState, useCallback } from 'react';
import { useDropzone,Accept } from 'react-dropzone';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { X } from 'lucide-react';

const testTypes = [
  { value: 'typing', label: 'Typing Test' },
  { value: 'docx', label: 'Docx Test' },
  { value: 'excel', label: 'Excel Test' },
  { value: 'word', label: 'Word Test' },
];

export default function AddTest() {
  const [testType, setTestType] = useState('typing');
  const [duration, setDuration] = useState('');
  const [typingText, setTypingText] = useState('');
  const [file, setFile] = useState<any>(null);

  const onDrop = useCallback((acceptedFiles:any) => {
    setFile(acceptedFiles[0]);
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'image/*': []
    },
    multiple: false,
  });

  const handleSubmit = (e:React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log({
      testType,
      duration: parseInt(duration, 10),
      typingText: testType === 'typing' ? typingText : undefined,
      file: testType !== 'typing' ? file : undefined,
    });
    
  };

  const removeFile = () => {
    setFile(null);
  };

  return (
    <div className="w-full mx-auto p-4 space-y-6">
      <h1 className="text-3xl font-bold mb-8">Add New Test</h1>
      <Card>
        <CardContent className="p-6 w-full">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="testType">Test Type</Label>
              <Select value={testType} onValueChange={setTestType}>
                <SelectTrigger>
                  <SelectValue placeholder="Select test type" />
                </SelectTrigger>
                <SelectContent>
                  {testTypes.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="duration">Duration (minutes)</Label>
              <Input
                type="number"
                id="duration"
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                placeholder="Enter test duration"
                min="1"
                required
              />
            </div>

            {testType === 'typing' ? (
              <div className="space-y-2">
                <Label htmlFor="typingText">Typing Text</Label>
                <Textarea
                  id="typingText"
                  value={typingText}
                  onChange={(e) => setTypingText(e.target.value)}
                  placeholder="Enter the text for the typing test"
                  rows={6}
                  required
                />
              </div>
            ) : (
              <div className="space-y-2">
                <Label>Upload File</Label>
                <div
                  {...getRootProps()}
                  className={`h-32 border-2 border-dashed rounded-md p-4 text-center cursor-pointer transition-colors ${
                    isDragActive ? 'border-primary bg-primary/10' : 'border-gray-300 hover:border-primary'
                  }`}
                >
                  <input {...getInputProps()} />
                  {isDragActive ? (
                    <p>Drop the file here ...</p>
                  ) : (
                    <p>Drag 'n' drop a file here, or click to select a file</p>
                  )}
                </div>
                {file && (
                  <Alert className="mt-2">
                    <AlertDescription className="flex items-center justify-between">
                      <span>{file.name}</span>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={removeFile}
                        className="p-0 h-auto"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            )}

            <Button type="submit" className="w-full mt-4">
              Add Test
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}